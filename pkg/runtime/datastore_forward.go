package runtime

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"

	"github.com/duso-org/duso/pkg/core"
	"github.com/duso-org/duso/pkg/script"
)

// Write forwarding: a follower sends a mutation to the leader, the leader
// applies it and replies with the result, and the follower returns that result
// once the corresponding frame has arrived on the normal stream.
//
// The follower never applies the write itself. Local state is only ever mutated
// by the apply loop, in sequence order — applying on reply would land the write
// at a position in the sequence that does not exist, and two replicas writing
// the same key would diverge permanently.
//
// Ordering falls out of using one socket: the leader sends the frame before the
// reply, the apply loop reads them in that order, and the caller is unblocked by
// the apply loop after it applies. So a forwarded write is visible to the very
// next local read.

const (
	// replForwardTimeout bounds a non-blocking forwarded write. Blocking ops
	// (shift_wait, pop_wait) carry the caller's own timeout instead.
	replForwardTimeout = 30 * time.Second

	// replMaxLeaderWaiters caps the goroutines one follower can park on a leader
	// with blocking ops, so a misbehaving replica cannot exhaust it.
	replMaxLeaderWaiters = 256
)

// Leader-to-follower message tags. An empty message is still a heartbeat.
const (
	replMsgFrames uint8 = 0x01 // concatenated WAL frames
	replMsgReply  uint8 = 0x02 // result of a forwarded write
)

// Request-only opcodes, above the WAL's control range so they can never collide
// with a real record opcode.
const (
	opReqShiftWait walOp = 0xC0
	opReqPopWait   walOp = 0xC1
)

// ---------------------------------------------------------------------------
// Wire encoding
// ---------------------------------------------------------------------------

type replRequest struct {
	ID       uint64
	Op       walOp
	Key      string
	Key2     string
	Num      float64
	Value    any
	HasValue bool
	Timeout  float64 // seconds; 0 means none
}

type replReply struct {
	ID       uint64
	Seq      uint64 // frame this write produced; 0 when it produced none
	Err      string
	Value    any
	HasValue bool
}

func encodeReplRequest(r *replRequest) ([]byte, error) {
	buf := make([]byte, 0, 64)
	buf = binary.AppendUvarint(buf, r.ID)
	buf = append(buf, byte(r.Op))
	buf = binary.LittleEndian.AppendUint64(buf, uint64(int64(r.Timeout*1e9)))
	buf = appendWALString(buf, r.Key)
	buf = appendWALString(buf, r.Key2)
	buf = binary.LittleEndian.AppendUint64(buf, math.Float64bits(r.Num))
	if r.HasValue {
		buf = append(buf, 1)
		var err error
		if buf, err = script.EncodeValue(buf, r.Value); err != nil {
			return nil, err
		}
	} else {
		buf = append(buf, 0)
	}
	return buf, nil
}

func decodeReplRequest(b []byte) (*replRequest, error) {
	r := &replRequest{}
	id, n := binary.Uvarint(b)
	if n <= 0 {
		return nil, fmt.Errorf("malformed request id")
	}
	r.ID, b = id, b[n:]
	if len(b) < 1 {
		return nil, fmt.Errorf("truncated request")
	}
	r.Op, b = walOp(b[0]), b[1:]
	if len(b) < 8 {
		return nil, fmt.Errorf("truncated request timeout")
	}
	r.Timeout = float64(int64(binary.LittleEndian.Uint64(b))) / 1e9
	b = b[8:]

	var err error
	if r.Key, b, err = readWALString(b); err != nil {
		return nil, err
	}
	if r.Key2, b, err = readWALString(b); err != nil {
		return nil, err
	}
	if len(b) < 9 {
		return nil, fmt.Errorf("truncated request body")
	}
	r.Num = math.Float64frombits(binary.LittleEndian.Uint64(b))
	b = b[8:]
	r.HasValue, b = b[0] == 1, b[1:]
	if r.HasValue {
		if r.Value, _, err = script.DecodeValue(b); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func encodeReplReply(r *replReply) ([]byte, error) {
	buf := make([]byte, 0, 64)
	buf = binary.AppendUvarint(buf, r.ID)
	buf = binary.AppendUvarint(buf, r.Seq)
	buf = appendWALString(buf, r.Err)
	if r.HasValue {
		buf = append(buf, 1)
		var err error
		if buf, err = script.EncodeValue(buf, r.Value); err != nil {
			return nil, err
		}
	} else {
		buf = append(buf, 0)
	}
	return buf, nil
}

func decodeReplReply(b []byte) (*replReply, error) {
	r := &replReply{}
	id, n := binary.Uvarint(b)
	if n <= 0 {
		return nil, fmt.Errorf("malformed reply id")
	}
	r.ID, b = id, b[n:]
	seq, n := binary.Uvarint(b)
	if n <= 0 {
		return nil, fmt.Errorf("malformed reply seq")
	}
	r.Seq, b = seq, b[n:]

	var err error
	if r.Err, b, err = readWALString(b); err != nil {
		return nil, err
	}
	if len(b) < 1 {
		return nil, fmt.Errorf("truncated reply")
	}
	r.HasValue, b = b[0] == 1, b[1:]
	if r.HasValue {
		if r.Value, _, err = script.DecodeValue(b); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// ---------------------------------------------------------------------------
// Follower side
// ---------------------------------------------------------------------------

// replConn is one live connection to the leader, shared by every script
// goroutine that wants to forward a write.
type replConn struct {
	ws      *websocket.Conn
	writeMu sync.Mutex // websockets allow concurrent read+write, but not concurrent writes

	nextID  atomic.Uint64
	pendMu  sync.Mutex
	pending map[uint64]chan *replReply

	closeOnce sync.Once
	closed    chan struct{}
}

func newReplConn(ws *websocket.Conn) *replConn {
	return &replConn{
		ws:      ws,
		pending: make(map[uint64]chan *replReply),
		closed:  make(chan struct{}),
	}
}

// shutdown wakes every caller still waiting on this connection. Their writes are
// of unknown outcome, which is what they are told.
func (c *replConn) shutdown() {
	c.closeOnce.Do(func() { close(c.closed) })
}

// deliver hands a reply to its waiter. Called from the apply loop, after the
// frame this reply refers to has already been applied.
func (c *replConn) deliver(r *replReply) {
	c.pendMu.Lock()
	ch, ok := c.pending[r.ID]
	delete(c.pending, r.ID)
	c.pendMu.Unlock()
	if ok {
		ch <- r
	}
}

func (c *replConn) register() (uint64, chan *replReply) {
	id := c.nextID.Add(1)
	ch := make(chan *replReply, 1)
	c.pendMu.Lock()
	c.pending[id] = ch
	c.pendMu.Unlock()
	return id, ch
}

func (c *replConn) abandon(id uint64) {
	c.pendMu.Lock()
	delete(c.pending, id)
	c.pendMu.Unlock()
}

// isFollower reports whether writes to this store belong to another machine.
func (ds *DatastoreValue) isFollower() bool {
	return ds.repl != nil && ds.repl.role == replRoleFollower
}

// replForward sends one mutation to the leader and returns the leader's result.
//
// Callers must not hold dataMutex: the reply is delivered by the apply loop,
// which needs that lock to apply the frame first.
func (ds *DatastoreValue) replForward(procCtx context.Context, req *replRequest, timeout time.Duration) (any, error) {
	repl := ds.repl

	// Connection first: a follower that has never connected has not been granted
	// anything either, and "not connected" is the more useful of the two.
	conn := repl.conn.Load()
	if conn == nil {
		return nil, fmt.Errorf("datastore(%q): not connected to leader %s — the write was not applied",
			ds.namespace, repl.leaderURL)
	}
	if !repl.canWrite.Load() {
		return nil, fmt.Errorf("datastore(%q): this replica has read-only access to %s — the write was not applied",
			ds.namespace, repl.leaderURL)
	}

	id, ch := conn.register()
	req.ID = id

	body, err := encodeReplRequest(req)
	if err != nil {
		conn.abandon(id)
		return nil, fmt.Errorf("datastore(%q): cannot forward write: %v", ds.namespace, err)
	}

	conn.writeMu.Lock()
	err = websocket.Message.Send(conn.ws, body)
	conn.writeMu.Unlock()
	if err != nil {
		conn.abandon(id)
		return nil, fmt.Errorf("datastore(%q): lost the leader while forwarding — the write was not applied: %v",
			ds.namespace, err)
	}

	if procCtx == nil {
		procCtx = context.Background()
	}
	var deadline <-chan time.Time
	if timeout > 0 {
		t := time.NewTimer(timeout)
		defer t.Stop()
		deadline = t.C
	}

	select {
	case reply := <-ch:
		if reply.Err != "" {
			return nil, fmt.Errorf("%s", reply.Err)
		}
		return reply.Value, nil

	case <-conn.closed:
		conn.abandon(id)
		// The leader may or may not have applied this before the socket died, and
		// nothing on this side can tell which. Saying so beats implying either.
		return nil, fmt.Errorf("datastore(%q): connection to leader %s ended while the write was in flight — it may or may not have been applied",
			ds.namespace, repl.leaderURL)

	case <-procCtx.Done():
		conn.abandon(id)
		return nil, fmt.Errorf("datastore(%q): process ended while a write was in flight — it may or may not have been applied", ds.namespace)

	case <-deadline:
		conn.abandon(id)
		return nil, fmt.Errorf("datastore(%q): leader %s did not answer within %s — the write may or may not have been applied",
			ds.namespace, repl.leaderURL, timeout)
	}
}

// forwardSimple is the common shape: a mutation with no timeout, run against the
// calling process's context so kill(pid) wakes it.
func (ds *DatastoreValue) forwardSimple(op walOp, key, key2 string, num float64, value any, hasValue bool) (any, error) {
	return ds.replForward(context.Background(), &replRequest{
		Op: op, Key: key, Key2: key2, Num: num, Value: value, HasValue: hasValue,
	}, replForwardTimeout)
}

// ---------------------------------------------------------------------------
// Leader side
// ---------------------------------------------------------------------------

// replExecute runs one forwarded request against the leader's own store, using
// the ordinary mutation methods so a forwarded write is indistinguishable from a
// local one — same locking, same WAL record, same broadcast to every follower.
func (ds *DatastoreValue) replExecute(req *replRequest) (any, error) {
	switch req.Op {
	case opSet:
		return nil, ds.Set(req.Key, req.Value)
	case opSetOnce:
		return ds.SetOnce(req.Key, req.Value)
	case opSwap:
		return ds.Swap(req.Key, req.Value)
	case opUpdate:
		return ds.Update(req.Key, req.Value)
	case opIncr:
		return ds.Increment(req.Key, req.Num)
	case opPush:
		return ds.Push(req.Key, req.Value)
	case opUnshift:
		return ds.Unshift(req.Key, req.Value)
	case opShift:
		return ds.Shift(req.Key)
	case opPop:
		return ds.Pop(req.Key)
	case opDelete:
		return ds.Delete(req.Key)
	case opClear:
		return nil, ds.Clear()
	case opRename:
		return nil, ds.Rename(req.Key, req.Key2)
	case opExpire:
		return nil, ds.Expire(req.Key, req.Num)
	case opReqShiftWait:
		return ds.ShiftWait(context.Background(), req.Key, time.Duration(req.Timeout*float64(time.Second)))
	case opReqPopWait:
		return ds.PopWait(context.Background(), req.Key, time.Duration(req.Timeout*float64(time.Second)))
	default:
		return nil, fmt.Errorf("unknown forwarded operation %#x", byte(req.Op))
	}
}

// replConnState is the leader's per-follower request/reply bookkeeping.
type replConnState struct {
	canWrite bool
	mu       sync.Mutex
	replies  []*replReply
	waiters  atomic.Int64
}

// queue holds a reply until the sender has shipped the frame it refers to.
func (s *replConnState) queue(r *replReply) {
	s.mu.Lock()
	s.replies = append(s.replies, r)
	s.mu.Unlock()
}

// take returns the replies that are safe to send now: those whose frame has
// already gone out, plus rejections, which produced no frame at all.
func (s *replConnState) take(cursor uint64) []*replReply {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.replies) == 0 {
		return nil
	}
	var ready, held []*replReply
	for _, r := range s.replies {
		if r.Seq == 0 || r.Seq <= cursor {
			ready = append(ready, r)
		} else {
			held = append(held, r)
		}
	}
	s.replies = held
	return ready
}

// replServeRequests reads forwarded writes from one follower until the socket
// closes. Each request runs on its own goroutine so a blocking shift_wait does
// not stall the writes queued behind it.
func (l *replListener) replServeRequests(ws *websocket.Conn, ds *DatastoreValue, state *replConnState, done chan struct{}) {
	repl := ds.repl
	for {
		var body []byte
		if err := websocket.Message.Receive(ws, &body); err != nil {
			return // socket closed; serve() notices via its own write failing
		}
		if len(body) == 0 {
			continue
		}

		req, err := decodeReplRequest(body)
		if err != nil {
			state.queue(&replReply{Err: fmt.Sprintf("malformed forwarded request: %v", err)})
			repl.ring.wake()
			continue
		}

		// Checked here as well as on the follower: the follower's own check is a
		// convenience that saves a round trip, this one is the actual boundary.
		if !state.canWrite {
			state.queue(&replReply{ID: req.ID,
				Err: fmt.Sprintf("datastore(%q): this replica has read-only access", ds.namespace)})
			repl.ring.wake()
			continue
		}

		if state.waiters.Load() >= replMaxLeaderWaiters {
			state.queue(&replReply{ID: req.ID,
				Err: fmt.Sprintf("leader is already holding %d in-flight writes for this follower", replMaxLeaderWaiters)})
			repl.ring.wake()
			continue
		}

		state.waiters.Add(1)
		go func() {
			defer core.RecoverPanic("datastore_replication_execute")
			defer state.waiters.Add(-1)

			seqBefore := ds.walSeq.Load()
			result, err := ds.replExecute(req)

			reply := &replReply{ID: req.ID}
			if err != nil {
				reply.Err = err.Error()
			} else {
				reply.Value, reply.HasValue = result, true
				// Only claim a sequence if this write actually produced a record;
				// a no-op shift on an empty array does not, and holding its reply
				// for a frame that never comes would hang the caller.
				if after := ds.walSeq.Load(); after > seqBefore {
					reply.Seq = after
				}
			}

			select {
			case <-done:
				return
			default:
			}
			state.queue(reply)
			repl.ring.wake()
		}()
	}
}

// sendReplies ships every reply whose frame has already been sent.
func replSendReplies(ws *websocket.Conn, state *replConnState, cursor uint64) error {
	for _, r := range state.take(cursor) {
		body, err := encodeReplReply(r)
		if err != nil {
			return err
		}
		if err := websocket.Message.Send(ws, append([]byte{replMsgReply}, body...)); err != nil {
			return err
		}
	}
	return nil
}
