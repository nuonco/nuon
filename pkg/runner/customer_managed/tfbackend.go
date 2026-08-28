package customermanaged

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
)

type TFBackend struct {
	store  statestore.Store
	server *http.Server
	ln     net.Listener
}

// NewTFBackend starts the loopback state backend. Terraform bakes the
// backend URL (including the port) into every saved plan, so a resumed
// invocation must bind the exact port a previous invocation used or chained
// apply-plan steps would be rejected as "created for a different backend".
// The chosen port is persisted at portFile (pass "" to always pick an
// ephemeral port, e.g. in tests).
func NewTFBackend(store statestore.Store, portFile string) (*TFBackend, error) {
	addr := "127.0.0.1:0"
	if portFile != "" {
		if b, err := os.ReadFile(portFile); err == nil {
			addr = "127.0.0.1:" + strings.TrimSpace(string(b))
		}
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		if addr != "127.0.0.1:0" {
			return nil, fmt.Errorf("terraform backend port %s (persisted in %s) is unavailable: %w — plans created by previous invocations are bound to this port; free it and retry", addr, portFile, err)
		}
		return nil, err
	}
	if portFile != "" {
		_, port, splitErr := net.SplitHostPort(ln.Addr().String())
		if splitErr != nil {
			ln.Close()
			return nil, splitErr
		}
		if writeErr := os.WriteFile(portFile, []byte(port+"\n"), 0o600); writeErr != nil {
			ln.Close()
			return nil, fmt.Errorf("persist terraform backend port: %w", writeErr)
		}
	}
	backend := &TFBackend{store: store, ln: ln}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/terraform-backend", backend.state)
	mux.HandleFunc("/v1/terraform-workspaces/", backend.lock)
	backend.server = &http.Server{Handler: mux}
	go func() { _ = backend.server.Serve(ln) }()
	return backend, nil
}

func (b *TFBackend) Addr() string { return b.ln.Addr().String() }
func (b *TFBackend) Close() error {
	err := b.server.Close()
	// server.Close only closes listeners Serve has already registered; close
	// ours directly so the port is released even if Serve hasn't started yet.
	if lnErr := b.ln.Close(); lnErr != nil && !errors.Is(lnErr, net.ErrClosed) && err == nil {
		err = lnErr
	}
	return err
}

func (b *TFBackend) state(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("workspace_id")
	if id == "" {
		http.Error(w, "workspace_id is required", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		state, ok, err := b.store.GetTFState(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok || len(state) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(state)
	case http.MethodPost:
		state, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := b.store.PutTFState(id, state); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (b *TFBackend) lock(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/terraform-workspaces/"), "/")
	if len(parts) != 2 || parts[0] == "" || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch parts[1] {
	case "lock":
		err = b.store.LockTF(parts[0], body)
		var conflict *statestore.LockConflictError
		if errors.As(err, &conflict) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusLocked)
			_, _ = w.Write(conflict.Existing)
			return
		}
	case "unlock":
		err = b.store.UnlockTF(parts[0])
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("terraform lock operation: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
