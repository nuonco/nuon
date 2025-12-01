package handlers

import (
	"sync"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Global map to store open documents
var openDocuments = make(map[protocol.DocumentUri]string)

// Mutex to protect concurrent access to openDocuments map
var openDocumentsMutex sync.RWMutex
