package parse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAppBranchConfig_RequiresName(t *testing.T) {
	_, err := ParseAppBranchConfig(strings.NewReader(`
connected_repo = { repo = "acme/platform", directory = ".", branch = "main" }
`))
	if err == nil || !strings.Contains(err.Error(), "name") {
		t.Fatalf("expected missing name error, got %v", err)
	}
}

func TestParseAppBranchConfig_ValidStandalone(t *testing.T) {
	cfg, err := ParseAppBranchConfig(strings.NewReader(`
name = "staging"

[connected_repo]
repo = "acme/platform"
directory = "."
branch = "staging"

[[install_groups]]
name = "canary"
order = 0
label_selector = { tier = "canary" }
auto_approve_on_policies_passing = true
`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "staging" {
		t.Fatalf("name = %q", cfg.Name)
	}
	if len(cfg.InstallGroups) != 1 || cfg.InstallGroups[0].LabelSelector["tier"] != "canary" {
		t.Fatalf("install groups = %+v", cfg.InstallGroups)
	}
}

func TestParseAppBranchConfig_InvalidRunModeListsValidModes(t *testing.T) {
	_, err := ParseAppBranchConfig(strings.NewReader(`
name = "staging"

[run]
mode = "sometimes"
`))
	if err == nil {
		t.Fatal("expected invalid run mode error")
	}
	for _, want := range []string{"sometimes", "push", "on_tag", "on_github_label", "manual_only"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to contain %q, got %v", want, err)
		}
	}
}

func TestParseAppBranchConfig_AcceptsLegacyTagMode(t *testing.T) {
	_, err := ParseAppBranchConfig(strings.NewReader(`
name = "staging"

[run]
mode = "on_tag_prefix"
tag_prefix = "release/"
`))
	if err != nil {
		t.Fatalf("expected legacy run mode to remain valid, got %v", err)
	}
}

func TestParseAppBranchConfig_InvalidPreviewModeListsValidModes(t *testing.T) {
	_, err := ParseAppBranchConfig(strings.NewReader(`
name = "staging"

[preview]
mode = "sometimes"
install_name = "example"
`))
	if err == nil {
		t.Fatal("expected invalid preview mode error")
	}
	for _, want := range []string{"sometimes", "none", "plan-only", "apply", "build-only"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to contain %q, got %v", want, err)
		}
	}
}

func TestParseAppBranchConfig_PreviewNoneNeedsNoInstall(t *testing.T) {
	cfg, err := ParseAppBranchConfig(strings.NewReader(`
name = "staging"

[preview]
mode = "none"
`))
	if err != nil {
		t.Fatalf("expected preview mode none to be valid, got %v", err)
	}
	if cfg.Preview == nil || cfg.Preview.Mode != "none" {
		t.Fatalf("expected preview mode none, got %#v", cfg.Preview)
	}
}

func TestParseAppBranchConfig_PreviewNoneRejectsInstall(t *testing.T) {
	_, err := ParseAppBranchConfig(strings.NewReader(`
name = "staging"

[preview]
mode = "none"
install_name = "example"
`))
	if err == nil || !strings.Contains(err.Error(), "mode none cannot set") {
		t.Fatalf("expected preview mode none target error, got %v", err)
	}
}

func TestParseAppBranchConfigDir_RejectsDuplicateNames(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "a.toml"), "name = \"main\"\n")
	mustWrite(t, filepath.Join(dir, "nested", "b.toml"), "name = \"main\"\n")

	_, err := ParseAppBranchConfigDir(dir)
	if err == nil || !strings.Contains(err.Error(), `branch "main" is already declared`) {
		t.Fatalf("expected duplicate name error, got %v", err)
	}
}

func TestParseAppBranchConfigDir_SortsAndSkipsNonToml(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "z.toml"), "name = \"zeta\"\n")
	mustWrite(t, filepath.Join(dir, "a.toml"), "name = \"alpha\"\n")
	mustWrite(t, filepath.Join(dir, "notes.md"), "name = \"ignored\"\n")

	files, err := ParseAppBranchConfigDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files", len(files))
	}
	if files[0].Config.Name != "alpha" || files[1].Config.Name != "zeta" {
		t.Fatalf("sorted names = %s, %s", files[0].Config.Name, files[1].Config.Name)
	}
}

func TestLoadAppBranchConfigs_FileVsDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.toml")
	mustWrite(t, path, "name = \"main\"\n")

	files, directory, err := LoadAppBranchConfigs(path)
	if err != nil || directory || len(files) != 1 || files[0].Config.Name != "main" {
		t.Fatalf("file mode: dir=%v files=%v err=%v", directory, files, err)
	}

	files, directory, err = LoadAppBranchConfigs(dir)
	if err != nil || !directory || len(files) != 1 {
		t.Fatalf("dir mode: dir=%v files=%v err=%v", directory, files, err)
	}
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
