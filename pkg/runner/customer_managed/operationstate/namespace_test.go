package operationstate

import (
	"context"
	"testing"
)

func TestNamespacedStateAndLegacyReadOverlay(t *testing.T) {
	ctx := context.Background()
	base := NewLocal(t.TempDir())
	runner := WithPrefix(base, RunnerNamespace)
	control := WithPrefix(base, ControlNamespace)

	if err := runner.Put(ctx, "future/artifact.json", []byte("runner")); err != nil {
		t.Fatal(err)
	}
	if err := control.Put(ctx, "dispatch/request.json", []byte("control")); err != nil {
		t.Fatal(err)
	}
	if err := base.Put(ctx, "legacy.json", []byte("legacy")); err != nil {
		t.Fatal(err)
	}

	read := ReadOverlay(runner, control, Legacy(base))
	for key, expected := range map[string]string{
		"future/artifact.json":  "runner",
		"dispatch/request.json": "control",
		"legacy.json":           "legacy",
	} {
		raw, found, err := read.Get(ctx, key)
		if err != nil || !found || string(raw) != expected {
			t.Fatalf("read %s: %q %v %v", key, raw, found, err)
		}
	}

	keys, err := read.List(ctx, "")
	if err != nil || len(keys) != 3 {
		t.Fatalf("overlay list: %#v %v", keys, err)
	}
	if err := read.Put(ctx, "unknown.json", nil); err == nil {
		t.Fatal("read overlay accepted an unowned write")
	}
}

func TestLegacyOperationsRemainReadableWithoutChangingWrites(t *testing.T) {
	ctx := context.Background()
	base := NewLocal(t.TempDir())
	if err := base.Put(ctx, "day2/catalog.json", []byte("legacy")); err != nil {
		t.Fatal(err)
	}
	state := WithLegacyOperationsRead(base)
	raw, found, err := state.Get(ctx, "operations/catalog.json")
	if err != nil || !found || string(raw) != "legacy" {
		t.Fatalf("legacy read: %q %v %v", raw, found, err)
	}
	if err := state.Put(ctx, "operations/bundle.json", []byte("new")); err != nil {
		t.Fatal(err)
	}
	if _, found, err := base.Get(ctx, "operations/bundle.json"); err != nil || !found {
		t.Fatalf("new operation write: %v %v", found, err)
	}
	if _, found, err := base.Get(ctx, "day2/bundle.json"); err != nil || found {
		t.Fatalf("legacy key was written: %v %v", found, err)
	}
}
