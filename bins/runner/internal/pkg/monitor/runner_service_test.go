package monitor

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestWriteNuonRunnerService(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, RunnerServiceName)
	restartPendingPath := filepath.Join(dir, "restart-pending")
	restartState := runnerServiceRestartState{PreviousInvocationID: "old-invocation"}

	if err := writeNuonRunnerService(path, restartPendingPath, []byte("first definition\n"), restartState); err != nil {
		t.Fatal(err)
	}
	assertRunnerServiceDefinition(t, path, "first definition\n", 0o644)
	assertRunnerServiceRestartState(t, restartPendingPath, restartState)

	if err := writeNuonRunnerService(path, restartPendingPath, []byte("second definition\n"), restartState); err != nil {
		t.Fatal(err)
	}
	assertRunnerServiceDefinition(t, path, "second definition\n", 0o644)
	assertRunnerServiceRestartState(t, restartPendingPath, restartState)

	if err := os.Remove(restartPendingPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeNuonRunnerService(path, restartPendingPath, []byte("second definition\n"), restartState); err != nil {
		t.Fatal(err)
	}
	assertRunnerServiceDefinition(t, path, "second definition\n", 0o644)
	assertRunnerServiceRestartState(t, restartPendingPath, restartState)
}

func TestWriteNuonRunnerServiceMarksRestartBeforeReplacingDefinition(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing", RunnerServiceName)
	restartPendingPath := filepath.Join(dir, "restart-pending")
	restartState := runnerServiceRestartState{}

	if err := writeNuonRunnerService(path, restartPendingPath, []byte("definition\n"), restartState); err == nil {
		t.Fatal("expected service definition write to fail")
	}
	assertRunnerServiceRestartState(t, restartPendingPath, restartState)
}

func TestWriteNuonRunnerServiceWithoutDirectoryWritePermission(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, RunnerServiceName)
	restartPendingPath := filepath.Join(t.TempDir(), "restart-pending")
	if err := os.WriteFile(path, []byte("first definition\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	if err := writeNuonRunnerService(path, restartPendingPath, []byte("second definition\n"), runnerServiceRestartState{}); err != nil {
		t.Fatal(err)
	}
	assertRunnerServiceDefinition(t, path, "second definition\n", 0o644)
}

func TestGetRunnerServiceRestartAction(t *testing.T) {
	now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
	requested := now.Add(-time.Minute)
	tests := map[string]struct {
		restartState runnerServiceRestartState
		serviceState runnerServiceState
		want         runnerServiceRestartAction
	}{
		"initial request": {
			restartState: runnerServiceRestartState{PreviousInvocationID: "old"},
			serviceState: runnerServiceState{ActiveState: "active", InvocationID: "old"},
			want:         runnerServiceRestartRequest,
		},
		"unrequested invocation change": {
			restartState: runnerServiceRestartState{PreviousInvocationID: "old"},
			serviceState: runnerServiceState{ActiveState: "active", InvocationID: "new"},
			want:         runnerServiceRestartRequest,
		},
		"new invocation active": {
			restartState: runnerServiceRestartState{PreviousInvocationID: "old", RestartRequestedAt: requested},
			serviceState: runnerServiceState{ActiveState: "active", InvocationID: "new"},
			want:         runnerServiceRestartComplete,
		},
		"slow startup": {
			restartState: runnerServiceRestartState{PreviousInvocationID: "old", RestartRequestedAt: now.Add(-time.Hour)},
			serviceState: runnerServiceState{ActiveState: "activating", InvocationID: "new"},
			want:         runnerServiceRestartWait,
		},
		"slow shutdown": {
			restartState: runnerServiceRestartState{PreviousInvocationID: "old", RestartRequestedAt: now.Add(-time.Hour)},
			serviceState: runnerServiceState{ActiveState: "deactivating", InvocationID: "old"},
			want:         runnerServiceRestartWait,
		},
		"failed before retry delay": {
			restartState: runnerServiceRestartState{PreviousInvocationID: "old", RestartRequestedAt: now.Add(-runnerServiceRestartRetryInterval / 2)},
			serviceState: runnerServiceState{ActiveState: "failed"},
			want:         runnerServiceRestartWait,
		},
		"failed after retry delay": {
			restartState: runnerServiceRestartState{PreviousInvocationID: "old", RestartRequestedAt: now.Add(-runnerServiceRestartRetryInterval)},
			serviceState: runnerServiceState{ActiveState: "failed"},
			want:         runnerServiceRestartRequest,
		},
		"old invocation before retry delay": {
			restartState: runnerServiceRestartState{PreviousInvocationID: "old", RestartRequestedAt: now.Add(-runnerServiceRestartRetryInterval / 2)},
			serviceState: runnerServiceState{ActiveState: "active", InvocationID: "old"},
			want:         runnerServiceRestartWait,
		},
		"old invocation after retry delay": {
			restartState: runnerServiceRestartState{PreviousInvocationID: "old", RestartRequestedAt: now.Add(-runnerServiceRestartRetryInterval)},
			serviceState: runnerServiceState{ActiveState: "active", InvocationID: "old"},
			want:         runnerServiceRestartRequest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := getRunnerServiceRestartAction(test.restartState, test.serviceState, now)
			if got != test.want {
				t.Fatalf("unexpected action: got %d, want %d", got, test.want)
			}
		})
	}
}

func TestRequestRunnerServiceOperationIsNonBlocking(t *testing.T) {
	dir := t.TempDir()
	argsPath := filepath.Join(dir, "args")
	systemctlPath := filepath.Join(dir, "systemctl")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$SYSTEMCTL_ARGS_PATH\"\n"
	if err := os.WriteFile(systemctlPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("SYSTEMCTL_ARGS_PATH", argsPath)

	if err := requestRunnerServiceOperation(t.Context(), "restart"); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Fields(string(contents))
	want := []string{"--system", "--no-block", "restart", RunnerServiceName}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected systemctl arguments: got %v, want %v", got, want)
	}
}

func assertRunnerServiceDefinition(t *testing.T, path string, want string, mode os.FileMode) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != want {
		t.Fatalf("unexpected service definition: %q", contents)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != mode {
		t.Fatalf("unexpected service definition permissions: %o", info.Mode().Perm())
	}
}

func assertRunnerServiceRestartState(t *testing.T, path string, want runnerServiceRestartState) {
	t.Helper()
	got, err := readRunnerServiceRestartState(path)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("runner service restart was not marked as pending")
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("unexpected runner service restart state: got %+v, want %+v", *got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected restart state permissions: %o", info.Mode().Perm())
	}
}
