package main

import (
	"github.com/powertoolsdev/mono/bins/lsp/handlers"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"

	_ "github.com/tliron/commonlog/simple"
)

const lsName = "Nuon Language Server"

var version string = "0.0.1"
var handler protocol.Handler

func main() {
	commonlog.Configure(2, nil)

	handler = protocol.Handler{
		Initialize:             initialize,
		Shutdown:               shutdown,
		TextDocumentCompletion: handlers.TextDocumentCompletion,
		TextDocumentDidOpen:    handlers.TextDocumentDidOpen,
		TextDocumentDidChange:  handlers.TextDocumentDidChange,
		TextDocumentDidClose:   handlers.TextDocumentDidClose,
		TextDocumentHover:      handlers.TextDocumentHover,
		TextDocumentDidSave:    handlers.TextDocumentDidSave,
	}

	server := server.NewServer(&handler, lsName, true)

	server.RunStdio()
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	commonlog.NewInfoMessage(0, "Initializing server...")

	return protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: protocol.TextDocumentSyncKindFull,
			HoverProvider:    true,
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"=", " "},
			},
		},
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}, nil
}

func shutdown(context *glsp.Context) error {
	return nil
}
