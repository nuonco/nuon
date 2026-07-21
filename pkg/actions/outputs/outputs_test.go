package outputs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLine(t *testing.T) {
	t.Run("key value", func(t *testing.T) {
		out, err := ParseLine("pod_count=3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out["pod_count"] != "3" {
			t.Fatalf("expected pod_count=3, got %v", out["pod_count"])
		}
	})

	t.Run("json object", func(t *testing.T) {
		out, err := ParseLine(`{"pod_count": 3, "region": "us-west-2"}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out["region"] != "us-west-2" {
			t.Fatalf("expected region us-west-2, got %v", out["region"])
		}
	})

	t.Run("json array rejected", func(t *testing.T) {
		if _, err := ParseLine(`["a", "b"]`); err == nil {
			t.Fatal("expected error for top-level json array")
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		if _, err := ParseLine("not-a-pair"); err == nil {
			t.Fatal("expected error for unsupported format")
		}
	})
}

func TestParseFile(t *testing.T) {
	dir := t.TempDir()

	t.Run("missing file is empty", func(t *testing.T) {
		out, err := ParseFile(filepath.Join(dir, "does-not-exist.json"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out) != 0 {
			t.Fatalf("expected empty map, got %v", out)
		}
	})

	t.Run("merges lines", func(t *testing.T) {
		path := filepath.Join(dir, Filename(0))
		contents := "pod_count=3\n\n{\"region\": \"us-west-2\"}\n"
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}

		out, err := ParseFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out["pod_count"] != "3" {
			t.Fatalf("expected pod_count=3, got %v", out["pod_count"])
		}
		if out["region"] != "us-west-2" {
			t.Fatalf("expected region us-west-2, got %v", out["region"])
		}
	})
}

func TestFilename(t *testing.T) {
	if got := Filename(2); got != "2.nuon-outputs.json" {
		t.Fatalf("expected 2.nuon-outputs.json, got %s", got)
	}
}
