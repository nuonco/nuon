package day2state

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
