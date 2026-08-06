package airgap

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/runner/airgap/statestore"
)

func TestTFBackendPortPersistsAcrossRestarts(t *testing.T) {
	dir := t.TempDir()
	store, err := statestore.NewDisk(dir)
	if err != nil {
		t.Fatal(err)
	}
	portFile := filepath.Join(dir, "tfbackend-port")

	first, err := NewTFBackend(store, portFile)
	if err != nil {
		t.Fatal(err)
	}
	addr := first.Addr()
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := NewTFBackend(store, portFile)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if second.Addr() != addr {
		t.Fatalf("expected restarted backend to reuse %s, got %s", addr, second.Addr())
	}
}

func TestTFBackendContract(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	backend, err := NewTFBackend(store, "")
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	base := "http://" + backend.Addr()
	response, err := http.Get(base + "/v1/terraform-backend?workspace_id=workspace")
	if err != nil || response.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected empty state response: %#v %v", response, err)
	}
	response.Body.Close()
	response, err = http.Post(base+"/v1/terraform-backend?workspace_id=workspace", "application/json", strings.NewReader(`{"version":4}`))
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected update response: %#v %v", response, err)
	}
	response.Body.Close()
	response, _ = http.Get(base + "/v1/terraform-backend?workspace_id=workspace")
	body, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if string(body) != `{"version":4}` {
		t.Fatalf("unexpected state %s", body)
	}
	lockURL := base + "/v1/terraform-workspaces/workspace/lock"
	response, _ = http.Post(lockURL, "application/json", strings.NewReader(`{"ID":"first"}`))
	response.Body.Close()
	response, _ = http.Post(lockURL, "application/json", strings.NewReader(`{"ID":"second"}`))
	body, _ = io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode != http.StatusLocked || string(body) != `{"ID":"first"}` {
		t.Fatalf("unexpected lock conflict: %d %s", response.StatusCode, body)
	}
}
