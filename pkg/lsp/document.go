package lsp

import (
	"sync"
)

// Document represents an open document in the LSP server
type Document struct {
	URI     string
	Version int
	Text    string
}

// DocumentManager manages open documents
type DocumentManager struct {
	documents map[string]*Document
	mu        sync.RWMutex
}

// NewDocumentManager creates a new document manager
func NewDocumentManager() *DocumentManager {
	return &DocumentManager{
		documents: make(map[string]*Document),
	}
}

// Open adds a new document or returns existing one
func (dm *DocumentManager) Open(uri string, version int, text string) *Document {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	doc := &Document{
		URI:     uri,
		Version: version,
		Text:    text,
	}
	dm.documents[uri] = doc
	return doc
}

// Update updates the content of an open document
func (dm *DocumentManager) Update(uri string, version int, text string) *Document {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	doc, exists := dm.documents[uri]
	if !exists {
		doc = &Document{
			URI: uri,
		}
		dm.documents[uri] = doc
	}

	doc.Version = version
	doc.Text = text
	return doc
}

// Close removes a document from the manager
func (dm *DocumentManager) Close(uri string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	delete(dm.documents, uri)
}

// Get retrieves a document by URI
func (dm *DocumentManager) Get(uri string) *Document {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.documents[uri]
}

// All returns all open documents
func (dm *DocumentManager) All() []*Document {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	docs := make([]*Document, 0, len(dm.documents))
	for _, doc := range dm.documents {
		docs = append(docs, doc)
	}
	return docs
}

// Count returns the number of open documents
func (dm *DocumentManager) Count() int {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return len(dm.documents)
}
