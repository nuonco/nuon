package config

import (
	"strings"
	"testing"
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
