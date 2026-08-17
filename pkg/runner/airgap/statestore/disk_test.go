package statestore

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDiskWriteStatusKeepsRunHistory(t *testing.T) {
	root := t.TempDir()
	store, err := NewDisk(root)
	if err != nil {
		t.Fatal(err)
	}
	status := &Status{InstallID: "install", RunID: "run-1", RunType: RunTypeInstall, Status: RunStatusInProgress}
	if err := store.WriteStatus(status); err != nil {
		t.Fatal(err)
	}
	status.Status = RunStatusFinished
	if err := store.WriteStatus(status); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(InstallRunStatusKey("run-1"))))
	if err != nil {
		t.Fatal(err)
	}
	var archived Status
	if err := json.Unmarshal(raw, &archived); err != nil {
		t.Fatal(err)
	}
	if archived.Status != RunStatusFinished {
		t.Fatalf("expected archived status to track the run, got %q", archived.Status)
	}
	first, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(InstallRunEventKey("run-1", 1))))
	if err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(InstallRunEventKey("run-1", 2))))
	if err != nil {
		t.Fatal(err)
	}
	var firstEvent, secondEvent StatusEvent
	if err := json.Unmarshal(first, &firstEvent); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(second, &secondEvent); err != nil {
		t.Fatal(err)
	}
	if firstEvent.Status.Status != RunStatusInProgress || secondEvent.Status.Status != RunStatusFinished {
		t.Fatalf("events are not immutable snapshots: %#v %#v", firstEvent, secondEvent)
	}
	archived.Status = RunStatusFailed
	projection, _ := json.Marshal(archived)
	if err := os.WriteFile(filepath.Join(root, "status.json"), projection, 0o600); err != nil {
		t.Fatal(err)
	}
	latest, err := store.ReadStatus()
	if err != nil || latest.Status != RunStatusFinished {
		t.Fatalf("latest event must win over projection: %#v %v", latest, err)
	}
}

func TestDiskTerraformStateAndLock(t *testing.T) {
	store, err := NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.PutTFState("workspace", []byte(`{"version":4}`)); err != nil {
		t.Fatal(err)
	}
	state, ok, err := store.GetTFState("workspace")
	if err != nil || !ok || string(state) != `{"version":4}` {
		t.Fatalf("unexpected state: %s %v %v", state, ok, err)
	}
	if err := store.LockTF("workspace", []byte(`{"ID":"first"}`)); err != nil {
		t.Fatal(err)
	}
	err = store.LockTF("workspace", []byte(`{"ID":"second"}`))
	var conflict *LockConflictError
	if !errors.As(err, &conflict) || string(conflict.Existing) != `{"ID":"first"}` {
		t.Fatalf("expected typed lock conflict, got %v", err)
	}
}

func TestDiskShowDocumentDoesNotOverwriteBackendState(t *testing.T) {
	store, err := NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.PutTFState("workspace", []byte(`{"version":4,"serial":1}`)); err != nil {
		t.Fatal(err)
	}
	if err := store.PutTFStateShow("workspace", []byte(`{"format_version":"1.0"}`)); err != nil {
		t.Fatal(err)
	}
	state, ok, err := store.GetTFState("workspace")
	if err != nil || !ok {
		t.Fatalf("state read failed: %v %v", ok, err)
	}
	if string(state) != `{"version":4,"serial":1}` {
		t.Fatalf("show document overwrote backend state: %s", state)
	}
}

func TestDiskReadResult(t *testing.T) {
	store, err := NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := store.ReadResult("step"); ok || err != nil {
		t.Fatalf("missing result should be (nil, false, nil), got %v %v", ok, err)
	}
	if err := store.WriteResult("step", map[string]string{"contents": "abc"}); err != nil {
		t.Fatal(err)
	}
	raw, ok, err := store.ReadResult("step")
	if err != nil || !ok {
		t.Fatalf("result read failed: %v %v", ok, err)
	}
	var decoded map[string]string
	if err := json.Unmarshal(raw, &decoded); err != nil || decoded["contents"] != "abc" {
		t.Fatalf("unexpected persisted result: %s %v", raw, err)
	}
}
