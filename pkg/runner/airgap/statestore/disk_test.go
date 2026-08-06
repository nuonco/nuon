package statestore

import (
	"encoding/json"
	"errors"
	"testing"
)

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
