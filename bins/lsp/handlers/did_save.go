package handlers

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TextDocumentDidSave(ctx *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	uri := params.TextDocument.URI
	log.Infof("💾 Document saved: %s", uri)

	openDocumentsMutex.RLock()
	text, ok := openDocuments[uri]
	openDocumentsMutex.RUnlock()
	if !ok {
		log.Warningf("⚠️  Document not found in cache for didSave: %s (may have been closed)", uri)
		if params.Text != nil {
			text = *params.Text
		} else {
			log.Errorf("❌ No document text available for didSave: %s", uri)
			return nil
		}
	}

	if params.Text != nil {
		text = *params.Text
		openDocumentsMutex.Lock()
		openDocuments[uri] = text
		openDocumentsMutex.Unlock()
		log.Debugf("✅ Updated document from save notification, new length: %d chars", len(text))
	}

	PublishDiagnostics(ctx, uri, text)

	return nil
}
