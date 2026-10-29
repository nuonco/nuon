package config

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"
)

func TestSourceArchiveDiff(t *testing.T) {
	head := NewSourceArchive()
	if err := head.AddFile("components/api/main.tf", []byte("resource \"x\" \"y\" {\n  size = 1\n}\n")); err != nil {
		t.Fatal(err)
	}
	if err := head.AddFile("charts/api/values.yaml", []byte("replicaCount: 3\n")); err != nil {
		t.Fatal(err)
	}
	if err := head.AddFile("nuon.toml", []byte("version = \"v2\"\n")); err != nil {
		t.Fatal(err)
	}

	base := NewSourceArchive()
	if err := base.AddFile("components/api/main.tf", []byte("resource \"x\" \"y\" {\n  size = 2\n}\n")); err != nil {
		t.Fatal(err)
	}
	if err := base.AddFile("charts/api/values.yaml", []byte("replicaCount: 3\n")); err != nil {
		t.Fatal(err)
	}
	if err := base.AddFile("components/legacy/manifest.yaml", []byte("kind: Deployment\n")); err != nil {
		t.Fatal(err)
	}

	d := head.Diff(base)

	if d.TotalFiles != 4 {
		t.Errorf("TotalFiles = %d, want 4", d.TotalFiles)
	}
	if d.Unchanged != 1 {
		t.Errorf("Unchanged = %d, want 1", d.Unchanged)
	}

	ops := map[string]string{}
	patches := map[string]string{}
	for _, f := range d.Files {
		ops[f.Path] = f.Op
		patches[f.Path] = f.Patch
	}

	if ops["charts/api/values.yaml"] != SourceFileOpUnchanged {
		t.Errorf("values.yaml op = %q, want unchanged", ops["charts/api/values.yaml"])
	}
	if ops["components/api/main.tf"] != SourceFileOpModified {
		t.Errorf("main.tf op = %q, want modified", ops["components/api/main.tf"])
	}
	if ops["nuon.toml"] != SourceFileOpAdded {
		t.Errorf("nuon.toml op = %q, want added", ops["nuon.toml"])
	}
	if ops["components/legacy/manifest.yaml"] != SourceFileOpRemoved {
		t.Errorf("manifest.yaml op = %q, want removed", ops["components/legacy/manifest.yaml"])
	}

	if !strings.Contains(patches["components/api/main.tf"], "-  size = 2") ||
		!strings.Contains(patches["components/api/main.tf"], "+  size = 1") {
		t.Errorf("main.tf patch missing expected lines:\n%s", patches["components/api/main.tf"])
	}
	if patches["charts/api/values.yaml"] != "" {
		t.Errorf("unchanged file should have no patch, got %q", patches["charts/api/values.yaml"])
	}

	// deterministic sort order
	last := ""
	for _, f := range d.Files {
		if f.Path < last {
			t.Errorf("files not sorted: %s after %s", f.Path, last)
		}
		last = f.Path
	}
}

func TestSourceArchiveDiffNilBaseAddsEverything(t *testing.T) {
	head := NewSourceArchive()
	if err := head.AddFile("nuon.toml", []byte("version = \"v2\"\n")); err != nil {
		t.Fatal(err)
	}

	d := head.Diff(nil)
	if d.TotalFiles != 1 || d.Unchanged != 0 {
		t.Fatalf("nil base diff: TotalFiles=%d Unchanged=%d, want 1/0", d.TotalFiles, d.Unchanged)
	}
	if d.Files[0].Op != SourceFileOpAdded {
		t.Errorf("op = %q, want added", d.Files[0].Op)
	}
}

func TestSourceArchiveDiffPatchTruncation(t *testing.T) {
	base := NewSourceArchive()
	if err := base.AddFile("big.txt", []byte(strings.Repeat("x\n", 10000))); err != nil {
		t.Fatal(err)
	}
	head := NewSourceArchive()
	if err := head.AddFile("big.txt", []byte(strings.Repeat("y\n", 10000))); err != nil {
		t.Fatal(err)
	}

	d := head.Diff(base)
	f := d.Files[0]
	if f.Op != SourceFileOpModified {
		t.Fatalf("op = %q, want modified", f.Op)
	}
	if !f.PatchTruncated {
		t.Error("expected patch_truncated")
	}
	if int64(len(f.Patch)) > MaxSourceDiffPatchBytes {
		t.Errorf("patch length %d exceeds cap %d", len(f.Patch), MaxSourceDiffPatchBytes)
	}
}

func TestSourceFileSHA256Stable(t *testing.T) {
	a := sourceFileSHA256("x")
	b := sourceFileSHA256("x")
	c := sourceFileSHA256("y")
	if a != b || a == c {
		t.Errorf("sha256 unstable or colliding: %s %s %s", a, b, c)
	}
}

func TestSourceArchiveDiffProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		names := drawDistinctStrings(t, hegel.FromRegex(`[a-z]{1,8}`, true), 0, 8)
		old, new := NewSourceArchive(), NewSourceArchive()
		oldFiles, newFiles := map[string]string{}, map[string]string{}
		expected := map[string]string{}
		unchanged := 0
		for _, name := range names {
			path := "config/" + name + ".toml"
			before, after := hegel.Draw(t, sourceArchiveText()), hegel.Draw(t, sourceArchiveText())
			if hegel.Draw(t, hegel.Booleans()) {
				after = before
			}
			wasPresent, isPresent := hegel.Draw(t, hegel.Booleans()), hegel.Draw(t, hegel.Booleans())
			if wasPresent {
				oldFiles[path] = before
				require.NoError(t, old.AddFile(path, []byte(before)))
			}
			if isPresent {
				newFiles[path] = after
				require.NoError(t, new.AddFile(path, []byte(after)))
			}
			switch {
			case !wasPresent && !isPresent:
				continue
			case !wasPresent:
				expected[path] = SourceFileOpAdded
			case !isPresent:
				expected[path] = SourceFileOpRemoved
			case before != after:
				expected[path] = SourceFileOpModified
			default:
				expected[path] = SourceFileOpUnchanged
				unchanged++
			}
		}
		require.Equal(t, oldFiles, old.Files)
		require.Equal(t, newFiles, new.Files)
		oldSnapshot, newSnapshot := maps.Clone(old.Files), maps.Clone(new.Files)
		forward, reverse := new.Diff(old), old.Diff(new)
		require.Equal(t, len(expected), forward.TotalFiles)
		require.Len(t, forward.Files, len(expected))
		require.Equal(t, unchanged, forward.Unchanged)
		require.Equal(t, unchanged, reverse.Unchanged)
		inverse := map[string]string{SourceFileOpAdded: SourceFileOpRemoved, SourceFileOpRemoved: SourceFileOpAdded, SourceFileOpModified: SourceFileOpModified, SourceFileOpUnchanged: SourceFileOpUnchanged}
		for i, file := range forward.Files {
			require.Equal(t, expected[file.Path], file.Op)
			require.Equal(t, file.Path, reverse.Files[i].Path)
			require.Equal(t, inverse[file.Op], reverse.Files[i].Op)
			assertSourceFileObservation(t, file, oldFiles, newFiles)
			assertSourceFileObservation(t, reverse.Files[i], newFiles, oldFiles)
			if i > 0 {
				require.Less(t, forward.Files[i-1].Path, file.Path)
			}
			if file.Op == SourceFileOpUnchanged {
				require.Empty(t, file.Patch)
				require.Equal(t, file.BeforeSHA256, file.AfterSHA256)
			} else if file.Op == SourceFileOpModified {
				require.NotEqual(t, file.BeforeSHA256, file.AfterSHA256)
			}
		}
		machine, err := json.Marshal(forward)
		require.NoError(t, err)
		permutedOld, permutedNew := NewSourceArchive(), NewSourceArchive()
		for _, path := range drawPermutation(t, slices.Sorted(maps.Keys(oldFiles))) {
			permutedOld.Files[path] = oldFiles[path]
		}
		for _, path := range drawPermutation(t, slices.Sorted(maps.Keys(newFiles))) {
			permutedNew.Files[path] = newFiles[path]
			permutedNew.Members["action:"+path] = path
		}
		permutedMachine, err := json.Marshal(permutedNew.Diff(permutedOld))
		require.NoError(t, err)
		require.Equal(t, string(machine), string(permutedMachine))
		require.Equal(t, forward, new.Diff(old))
		require.Equal(t, len(new.Files), new.Diff(new).Unchanged)
		require.Equal(t, oldSnapshot, old.Files)
		require.Equal(t, newSnapshot, new.Files)
	})
}

func assertSourceFileObservation(t require.TestingT, file SourceFileDiff, old, new map[string]string) {
	before, wasPresent := old[file.Path]
	after, isPresent := new[file.Path]
	require.True(t, wasPresent || isPresent)
	want := SourceFileOpUnchanged
	switch {
	case !wasPresent:
		want = SourceFileOpAdded
	case !isPresent:
		want = SourceFileOpRemoved
	case before != after:
		want = SourceFileOpModified
	}
	require.Equal(t, want, file.Op)
	for _, side := range []struct {
		contents string
		present  bool
		hash     string
		size     int64
	}{{before, wasPresent, file.BeforeSHA256, file.BeforeSize}, {after, isPresent, file.AfterSHA256, file.AfterSize}} {
		require.Equal(t, int64(len(side.contents)), side.size)
		if side.present {
			require.Equal(t, fmt.Sprintf("%x", sha256.Sum256([]byte(side.contents))), side.hash)
		} else {
			require.Empty(t, side.hash)
		}
	}
	require.LessOrEqual(t, len(file.Patch), MaxSourceDiffPatchBytes)
	require.True(t, utf8.ValidString(file.Patch))
	if want == SourceFileOpRemoved || want == SourceFileOpUnchanged {
		require.Empty(t, file.Patch)
		require.False(t, file.PatchTruncated)
		return
	}
	if !file.PatchTruncated {
		if before != after {
			require.NotEmpty(t, file.Patch)
		}
		rebuilt, err := applySourceArchivePatch(before, file.Patch)
		require.NoError(t, err, "%s: %q", file.Path, file.Patch)
		require.Equal(t, after, rebuilt, file.Path)
	}
}

func TestSourceArchiveDiffRenameAndEmptyPresence(t *testing.T) {
	for name, text := range map[string]string{"empty file": "", "nonempty file": "hello\n"} {
		t.Run(name, func(t *testing.T) {
			old := &SourceArchive{Files: map[string]string{"old.txt": text}}
			head := &SourceArchive{Files: map[string]string{"new.txt": text}}
			got := head.Diff(old)
			require.Equal(t, 2, got.TotalFiles)
			require.Zero(t, got.Unchanged)
			require.Equal(t, []string{"new.txt", "old.txt"}, []string{got.Files[0].Path, got.Files[1].Path})
			for _, file := range got.Files {
				assertSourceFileObservation(t, file, old.Files, head.Files)
			}
			added := head.Diff(nil)
			require.Equal(t, 1, added.TotalFiles)
			assertSourceFileObservation(t, added.Files[0], nil, head.Files)
			unchanged := head.Diff(head)
			require.Equal(t, 1, unchanged.Unchanged)
			assertSourceFileObservation(t, unchanged.Files[0], head.Files, head.Files)
		})
	}
}

func TestSourceArchiveDiffPatchCapBoundaries(t *testing.T) {
	for _, prefix := range []string{"x", "\u00e9", "\u20ac\U00010400", "--- header", "@@ -1 +1 @@"} {
		for _, ending := range []string{"\n", "\r\n", ""} {
			for _, modified := range []bool{false, true} {
				before := ""
				header := "--- a/cap.txt\n+++ b/cap.txt\n@@ -0,0 +1 @@\n"
				if modified {
					before = "old\n"
					header = "--- a/cap.txt\n+++ b/cap.txt\n@@ -1 +1 @@\n-old\n"
				}
				suffix := ending
				if ending == "" {
					suffix = "\n\\ No newline at end of file\n"
				}
				for _, delta := range []int{-2, -1, 0, 1, 2} {
					t.Run(fmt.Sprintf("%q/%q/modified=%t/delta=%d", prefix, ending, modified, delta), func(t *testing.T) {
						padding := MaxSourceDiffPatchBytes - len(header) - 1 - len(prefix) - len(suffix) + delta
						after := prefix + strings.Repeat("x", padding) + ending
						fullPatch := header + "+" + prefix + strings.Repeat("x", padding) + suffix
						old, head := NewSourceArchive(), NewSourceArchive()
						if modified {
							require.NoError(t, old.AddFile("cap.txt", []byte(before)))
						}
						require.NoError(t, head.AddFile("cap.txt", []byte(after)))
						file := head.Diff(old).Files[0]
						assertSourceFileObservation(t, file, old.Files, head.Files)
						require.Equal(t, delta > 0, file.PatchTruncated)
						if delta <= 0 {
							require.Equal(t, fullPatch, file.Patch)
						} else {
							require.True(t, strings.HasPrefix(fullPatch, file.Patch))
							require.True(t, strings.HasSuffix(file.Patch, "\n"))
						}
					})
				}
			}
		}
	}
}
