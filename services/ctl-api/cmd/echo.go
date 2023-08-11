package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/zap"
)

// EchoHandler is an http.Handler that copies its request body
// back to the response.
type EchoHandler struct {
	log *zap.Logger
}

// NewEchoHandler builds a new EchoHandler.
func NewEchoHandler(log *zap.Logger, cfg *internal.Config) *EchoHandler {
	return &EchoHandler{
		log: log,
	}
}

func (e *EchoHandler) Pattern() string {
	e.log.Info("echo handler pinged")
	return "/echo"
}

// ServeHTTP handles an HTTP request to the /echo endpoint.
func (*EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("abc"))

	if _, err := io.Copy(w, r.Body); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to handle request:", err)
	}
}
