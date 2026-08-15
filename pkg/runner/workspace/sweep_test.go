package workspace

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"go.uber.org/zap"
)

func TestSweepStale(t *testing.T) {
	t.Run("removes only workspace directories", func(t *testing.T) {
		root := t.TempDir()
		for _, dir := range []string{"workspace-exec-1", "workspace-exec-2", "action-images", "lost+found"} {
			if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}
		if err := os.WriteFile(filepath.Join(root, "workspace-not-a-dir"), []byte("x"), 0o600); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		removed, err := SweepStale(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		slices.Sort(removed)
		want := []string{"workspace-exec-1", "workspace-exec-2"}
		if !slices.Equal(removed, want) {
			t.Fatalf("expected %v, got %v", want, removed)
		}

		for _, keep := range []string{"action-images", "lost+found", "workspace-not-a-dir"} {
			if _, err := os.Stat(filepath.Join(root, keep)); err != nil {
				t.Fatalf("expected %s to survive: %v", keep, err)
			}
		}
	})

	t.Run("removes nested workspace content", func(t *testing.T) {
		root := t.TempDir()
		nested := filepath.Join(root, "workspace-exec-1", "deep", "deeper")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := os.WriteFile(filepath.Join(nested, "outputs"), []byte("k=v"), 0o600); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := SweepStale(root); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(filepath.Join(root, "workspace-exec-1")); !os.IsNotExist(err) {
			t.Fatalf("expected the workspace to be gone, got %v", err)
		}
	})

	t.Run("a missing root is not an error", func(t *testing.T) {
		removed, err := SweepStale(filepath.Join(t.TempDir(), "does-not-exist"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(removed) != 0 {
			t.Fatalf("expected nothing removed, got %v", removed)
		}
	})
}

func TestFilesystemType(t *testing.T) {
	// /proc/mounts only exists on linux; elsewhere an unknown type is reported
	// so callers treat it as fine rather than warning spuriously.
	if _, err := os.Stat("/proc/mounts"); err != nil {
		if got := FilesystemType("/tmp"); got != "" {
			t.Fatalf("expected an unknown filesystem without /proc/mounts, got %q", got)
		}
		return
	}

	if got := FilesystemType("/"); got == "" {
		t.Fatal("expected a filesystem type for /")
	}
}

// uncreatableDir returns a path that MkdirAll cannot create as any uid: a
// regular file sits where a parent directory would have to be, so it fails with
// ENOTDIR. Permission bits would not do, since tests run as root in CI and root
// bypasses the directory permission check.
func uncreatableDir(t *testing.T) string {
	t.Helper()

	blocker := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(blocker, "nested")
}

func TestResolveActionRoot(t *testing.T) {
	l := zap.NewNop()

	t.Run("prefers the first usable root", func(t *testing.T) {
		first := filepath.Join(t.TempDir(), "preferred")
		second := t.TempDir()

		if got := resolveActionRoot(l, []string{first, second}); got != first {
			t.Fatalf("expected %s, got %s", first, got)
		}
		if _, err := os.Stat(first); err != nil {
			t.Fatalf("expected the root to be created: %v", err)
		}
	})

	t.Run("falls through a root it cannot create", func(t *testing.T) {
		fallback := t.TempDir()

		if got := resolveActionRoot(l, []string{uncreatableDir(t), fallback}); got != fallback {
			t.Fatalf("expected the fallback %s, got %s", fallback, got)
		}
	})

	t.Run("falls back to the default when nothing is usable", func(t *testing.T) {
		got := resolveActionRoot(l, []string{uncreatableDir(t), uncreatableDir(t)})
		if got != DefaultTmpRootDir {
			t.Fatalf("expected %s, got %s", DefaultTmpRootDir, got)
		}
	})
}

func TestIsSubPath(t *testing.T) {
	for _, tc := range []struct {
		mountPoint string
		path       string
		want       bool
	}{
		{"/", "/opt/nuon/action-workspaces", true},
		{"/opt", "/opt/nuon/action-workspaces", true},
		{"/opt/nuon", "/opt/nuon", true},
		{"/opt/nuon/", "/opt/nuon/action-workspaces", true},
		{"/optional", "/opt/nuon", false},
		{"/var", "/opt/nuon", false},
	} {
		if got := isSubPath(tc.mountPoint, tc.path); got != tc.want {
			t.Fatalf("isSubPath(%q, %q) = %v, want %v", tc.mountPoint, tc.path, got, tc.want)
		}
	}
}
