package lsp

import (
	"sync"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// ASTCacheEntry represents a cached AST for a document
type ASTCacheEntry struct {
	URI         string
	Version     int
	AST         *script.Program
	Diagnostics []*Diagnostic
	Timestamp   time.Time
}

// ASTCache manages cached ASTs for documents
type ASTCache struct {
	cache map[string]*ASTCacheEntry
	mu    sync.RWMutex
	// Limit cache size to prevent unbounded memory growth
	maxEntries int
}

// NewASTCache creates a new AST cache
func NewASTCache(maxEntries int) *ASTCache {
	if maxEntries <= 0 {
		maxEntries = 100
	}
	return &ASTCache{
		cache:      make(map[string]*ASTCacheEntry),
		maxEntries: maxEntries,
	}
}

// Set stores an AST in the cache
func (ac *ASTCache) Set(uri string, version int, ast *script.Program, diags []*Diagnostic) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Evict oldest entry if at capacity
	if len(ac.cache) >= ac.maxEntries && ac.cache[uri] == nil {
		ac.evictOldest()
	}

	ac.cache[uri] = &ASTCacheEntry{
		URI:         uri,
		Version:     version,
		AST:         ast,
		Diagnostics: diags,
		Timestamp:   time.Now(),
	}
}

// Get retrieves an AST from the cache
func (ac *ASTCache) Get(uri string, version int) (*ASTCacheEntry, bool) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	entry, exists := ac.cache[uri]
	if !exists {
		return nil, false
	}

	// Return entry regardless of version (allows access to stale AST)
	return entry, entry.Version == version
}

// GetLatest retrieves the latest cached entry for a URI
func (ac *ASTCache) GetLatest(uri string) *ASTCacheEntry {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	return ac.cache[uri]
}

// Delete removes an entry from the cache
func (ac *ASTCache) Delete(uri string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	delete(ac.cache, uri)
}

// Clear removes all entries from the cache
func (ac *ASTCache) Clear() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.cache = make(map[string]*ASTCacheEntry)
}

// evictOldest removes the oldest entry from the cache
// Must be called with lock held
func (ac *ASTCache) evictOldest() {
	var oldest *ASTCacheEntry
	var oldestKey string

	for key, entry := range ac.cache {
		if oldest == nil || entry.Timestamp.Before(oldest.Timestamp) {
			oldest = entry
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(ac.cache, oldestKey)
	}
}
