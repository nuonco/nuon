package main

import (
	"flag"
	"fmt"
	"net/http"

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
	// Parse command line flags
	port := flag.Int("port", 0, "TCP port to listen on (0 = stdio mode for production)")
	healthPort := flag.Int("health-port", 0, "HTTP port for health checks (0 = disabled)")
	flag.Parse()

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

	// Start health check server if requested
	if *healthPort > 0 {
		go startHealthCheckServer(*healthPort)
	}

	if *port > 0 {
		// TCP mode for local development
		address := fmt.Sprintf("127.0.0.1:%d", *port)
		commonlog.NewInfoMessage(0, "Starting LSP server in TCP mode on %s", address)
		server.RunTCP(address)
	} else {
		// Stdio mode for production (used by VS Code extension)
		server.RunStdio()
	}
}

func startHealthCheckServer(port int) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	address := fmt.Sprintf("127.0.0.1:%d", port)
	commonlog.NewInfoMessage(0, "Health check server listening on %s", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		commonlog.NewErrorMessage(0, "Health check server failed: %s", err.Error())
	}
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
