package runtime

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// Process shutdown.
//
// Three things need to happen before duso exits, in this order:
//
//  1. WebSocket connections close, and scripts parked in ws.read() wake up.
//  2. HTTP servers stop accepting and let in-flight handlers finish.
//  3. Datastores sync their WAL and write a final snapshot.
//
// The order is the whole point. Handlers keep writing to datastores right up
// until the drain finishes, so flushing any earlier snapshots a store that is
// still being written.
//
// Both exit paths - a signal, and the script simply running out of statements -
// go through GracefulShutdown. sync.Once makes it safe to call from either, or
// both, or twice.

// drainTimeout bounds how long in-flight handlers get to finish. A handler that
// is streaming or long-polling will never finish on its own, so this is what
// stops a shutdown from hanging on one.
const drainTimeout = 30 * time.Second

var (
	shutdownOnce sync.Once
	shutdownDone = make(chan struct{}) // closed when shutdown has finished

	shutdownReqOnce sync.Once
	shutdownReq     = make(chan struct{}) // closed when shutdown has begun

	liveServerMutex sync.Mutex
	liveServers     []*HTTPServerValue
)

// ShutdownRequested is closed when shutdown begins. A blocked http_server waits
// on it instead of handling signals itself, so there is one signal owner.
func ShutdownRequested() <-chan struct{} { return shutdownReq }

// ShutdownComplete is closed once every server has drained and every datastore
// has flushed.
func ShutdownComplete() <-chan struct{} { return shutdownDone }

// registerLiveServer records a server that has successfully bound, so process
// shutdown knows to drain it. Registering at bind rather than at creation means
// a server that failed to start is never in the list.
func registerLiveServer(s *HTTPServerValue) {
	liveServerMutex.Lock()
	liveServers = append(liveServers, s)
	liveServerMutex.Unlock()
}

func unregisterLiveServer(s *HTTPServerValue) {
	liveServerMutex.Lock()
	for i, srv := range liveServers {
		if srv == s {
			liveServers = append(liveServers[:i], liveServers[i+1:]...)
			break
		}
	}
	liveServerMutex.Unlock()
}

// GracefulShutdown closes WebSockets, drains every running HTTP server, then
// flushes every datastore. It returns when all of that is done, and returns
// immediately on any call after the first has completed.
func GracefulShutdown() {
	shutdownOnce.Do(func() {
		defer close(shutdownDone)

		requestShutdown()

		// Drain first. SignalInterrupt below aborts sleep() (builtin_system.go:54)
		// as well as blocked WebSocket reads, so raising it any earlier cuts
		// short the very handlers the drain is meant to let finish.
		drainServers()

		// Then WebSockets. Go's Shutdown neither waits for nor closes hijacked
		// connections, so they are ours to deal with, and nothing above touched
		// them. Waking a handler parked in ws.read() may produce one last
		// datastore write, which the flush below still catches.
		SignalInterrupt()
		CloseAllConnections()

		// Last, once nothing can still be writing.
		ShutdownAllDatastores()
	})

	// A second caller arriving mid-shutdown waits for the first to finish
	// rather than racing it to os.Exit.
	<-shutdownDone
}

// requestShutdown wakes anything waiting on ShutdownRequested.
func requestShutdown() {
	shutdownReqOnce.Do(func() { close(shutdownReq) })
}

// drainServers stops every live server accepting new connections and waits for
// in-flight requests, up to drainTimeout. Servers drain concurrently: two
// servers should not cost two timeouts.
func drainServers() {
	liveServerMutex.Lock()
	servers := append([]*HTTPServerValue(nil), liveServers...)
	liveServerMutex.Unlock()

	if len(servers) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), drainTimeout)
	defer cancel()

	var wg sync.WaitGroup
	for _, s := range servers {
		if s.server == nil {
			continue
		}
		wg.Add(1)
		go func(srv *HTTPServerValue) {
			defer wg.Done()
			if err := srv.server.Shutdown(ctx); err != nil && err != context.Canceled {
				fmt.Fprintf(os.Stderr, "http_server (port=%d): shutdown: %v\n", srv.Port, err)
			}
		}(s)
	}
	wg.Wait()
}

// ShutdownAllDatastores flushes every registered datastore: WAL synced and
// closed, then a final snapshot for any store configured to persist.
//
// Without this, a store configured with persist but no persist_interval never
// writes its file at all - the snapshot ticker only runs when an interval is
// set, so the only other writer is an explicit save().
func ShutdownAllDatastores() {
	registryMutex.RLock()
	stores := make([]*DatastoreValue, 0, len(datastoreRegistry))
	for _, ds := range datastoreRegistry {
		stores = append(stores, ds)
	}
	registryMutex.RUnlock()

	for _, ds := range stores {
		if err := ds.Shutdown(); err != nil {
			// Report and keep going: one store failing to write its snapshot
			// is no reason to lose the others.
			fmt.Fprintf(os.Stderr, "datastore %q: shutdown: %v\n", ds.namespace, err)
		}
	}
}
