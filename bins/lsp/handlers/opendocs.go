package handlers

import (
	"sync"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

var openDocuments = make(map[protocol.DocumentUri]string)

var openDocumentsMutex sync.RWMutex
