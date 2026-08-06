package statestore

import (
	"encoding/json"
	"testing"
)

func TestDiskHealthRoundTrip(t *testing.T) {
	store, err := NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := store.ReadHealth(); ok || err != nil {
		t.Fatalf("missing health should be (nil, false, nil), got %v %v", ok, err)
	}
	if err := store.WriteHealth(map[string]string{"kind": "watch"}); err != nil {
		t.Fatal(err)
	}
	raw, ok, err := store.ReadHealth()
	if err != nil || !ok {
		t.Fatalf("health read failed: %v %v", ok, err)
	}
	var got map[string]string
	if err := json.Unmarshal(raw, &got); err != nil || got["kind"] != "watch" {
		t.Fatalf("unexpected health: %s %v", raw, err)
	}
}

func TestDiskHealthTransitionsAppend(t *testing.T) {
	store, err := NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := store.ReadHealthTransitions(); ok || err != nil {
		t.Fatalf("missing transitions should be (nil, false, nil), got %v %v", ok, err)
	}
	if err := store.AppendHealthTransitions(nil); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := store.ReadHealthTransitions(); ok {
		t.Fatal("empty append should not create the file")
	}
	if err := store.AppendHealthTransitions([]any{map[string]string{"to": "healthy"}}); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendHealthTransitions([]any{map[string]string{"to": "degraded"}, map[string]string{"to": "healthy"}}); err != nil {
		t.Fatal(err)
	}
	raw, ok, err := store.ReadHealthTransitions()
	if err != nil || !ok {
		t.Fatalf("transitions read failed: %v %v", ok, err)
	}
	var got []map[string]string
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0]["to"] != "healthy" || got[1]["to"] != "degraded" || got[2]["to"] != "healthy" {
		t.Fatalf("transitions should accumulate in order: %+v", got)
	}
}

func TestDiskHealthContextRoundTrip(t *testing.T) {
	store, err := NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := store.ReadHealthContext(); ok || err != nil {
		t.Fatalf("missing context should be (nil, false, nil), got %v %v", ok, err)
	}
	if err := store.WriteHealthContext(map[string]any{"sandbox_releases": []string{"ingress"}}); err != nil {
		t.Fatal(err)
	}
	raw, ok, err := store.ReadHealthContext()
	if err != nil || !ok {
		t.Fatalf("context read failed: %v %v", ok, err)
	}
	var got struct {
		SandboxReleases []string `json:"sandbox_releases"`
	}
	if err := json.Unmarshal(raw, &got); err != nil || len(got.SandboxReleases) != 1 {
		t.Fatalf("unexpected context: %s %v", raw, err)
	}
}
