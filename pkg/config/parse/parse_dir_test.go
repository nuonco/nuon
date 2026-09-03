package parse

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestParseDirErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("nonexistent directory", func(t *testing.T) {
		_, err := ParseDir(ctx, ParseConfig{
			Dirname: "/tmp/nonexistent-dir-abc123xyz",
		})
		require.Error(t, err)
		var errCfg config.ErrConfig
		require.ErrorAs(t, err, &errCfg)
		require.Contains(t, errCfg.Description, "does not exist")
	})

	t.Run("empty directory no toml files", func(t *testing.T) {
		dir := t.TempDir()

		_, err := ParseDir(ctx, ParseConfig{
			Dirname: dir,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to parse directory")
	})

	t.Run("invalid toml syntax", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "nuon.toml"), []byte("[[[bad toml"), 0644)
		require.NoError(t, err)

		_, err = ParseDir(ctx, ParseConfig{
			Dirname: dir,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to parse directory")
	})

	t.Run("valid toml missing version", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "nuon.toml"), []byte(`
[sandbox]
type = "unknown_type"
`), 0644)
		require.NoError(t, err)

		_, err = ParseDir(ctx, ParseConfig{
			Dirname: dir,
		})
		require.Error(t, err)
	})
}

func TestConfigDirIncludesTriggers(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "components"), 0755))
	files := map[string]string{
		"metadata.toml": `version = "v2"
display_name = "Event parser test"
`,
		"sandbox.toml": `terraform_version = "1.11.3"
[public_repo]
repo = "nuonco/aws-eks-sandbox"
directory = "."
branch = "main"
`,
		"runner.toml": `runner_type = "aws"
helm_driver = "configmap"
init_script_url = "https://example.com/init.sh"
`,
		"triggers.toml": `
[[rules]]
name = "push-image"
trigger = "registry"

[[rules.filters]]
path = "$.tag"
op = "suffix"
value = ":latest"

[rules.target]
type = "app_branch_run"
app_branch = "main"
`,
	}
	for name, contents := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0644))
	}

	cfg, err := ParseDir(context.Background(), ParseConfig{
		Dirname:       dir,
		FileProcessor: func(_ string, obj map[string]any) map[string]any { return obj },
	})
	require.NoError(t, err)
	require.Len(t, cfg.Triggers.Rules, 1)
	require.Equal(t, "push-image", cfg.Triggers.Rules[0].Name)
	require.Equal(t, "registry", cfg.Triggers.Rules[0].Trigger)
	require.Equal(t, "$.tag", cfg.Triggers.Rules[0].Filters[0].Path)
	require.Equal(t, "main", cfg.Triggers.Rules[0].Target.AppBranch)
}

func TestParseDirWithSourceCapturesConsumedSources(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "components"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "actions"), 0o755))

	metadata := `# preserve this authored comment
version = "v2"
description = "./not-a-get-field.txt"
readme = "./README.md"

[customer_managed]
runner_image_url = "registry.example.com/runner"
runner_image_tag = "v1"

[[customer_managed.platforms]]
target = "linux/amd64"
portal_binary_url = "https://artifacts.example.com/portal"
runner_binary_url = "https://artifacts.example.com/runner"
`
	action := `name = "hello"
timeout = "10s"

[[triggers]]
type = "manual"

[[steps]]
name = "say-hello"
command = "sh"
inline_contents = "./hello.sh"
`
	files := map[string]string{
		"metadata.toml":            metadata,
		"README.md":                "# Captured readme\n",
		"not-a-get-field.txt":      "must not be captured\n",
		"unreferenced.txt":         "must not be captured\n",
		"actions/hello.toml":       action,
		"actions/hello.sh":         "echo hello\n",
		"actions/unreferenced.txt": "must not be captured\n",
		"sandbox.toml": `terraform_version = "1.11.3"
[public_repo]
repo = "nuonco/aws-eks-sandbox"
directory = "."
branch = "main"
`,
		"runner.toml": `runner_type = "aws"
helm_driver = "configmap"
init_script_url = "https://example.com/init.sh"
`,
	}
	for name, contents := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644))
	}

	result, err := ParseDirWithSource(context.Background(), ParseConfig{
		Dirname:       dir,
		FileProcessor: func(_ string, obj map[string]any) map[string]any { return obj },
	})
	require.NoError(t, err)
	require.Equal(t, 3, result.Source.SchemaVersion)
	require.Equal(t, metadata, result.Source.Files["metadata.toml"])
	require.Equal(t, action, result.Source.Files["actions/hello.toml"])
	require.Equal(t, "# Captured readme\n", result.Source.Files["README.md"])
	require.Equal(t, "echo hello\n", result.Source.Files["actions/hello.sh"])
	require.NotContains(t, result.Source.Files, "not-a-get-field.txt")
	require.NotContains(t, result.Source.Files, "unreferenced.txt")
	require.NotContains(t, result.Source.Files, "actions/unreferenced.txt")
	require.Equal(t, "metadata.toml", result.Source.Members["metadata:metadata"])
	require.Equal(t, "actions/hello.toml", result.Source.Members["action:hello"])
	require.Equal(t, "sandbox.toml", result.Source.Members["sandbox:sandbox"])
	require.Equal(t, "runner.toml", result.Source.Members["runner:runner"])
	require.NoError(t, result.Source.ReindexMembers())
}

func TestParseDirWithSourceDoesNotApplyArchiveLimitsToStandardApps(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "components"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "metadata.toml"),
		[]byte("version = \"v2\"\n# "+strings.Repeat("x", maxSourceArchiveTestFileBytes)),
		0o644,
	))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sandbox.toml"), []byte(`terraform_version = "1.11.3"
[public_repo]
repo = "nuonco/aws-eks-sandbox"
directory = "."
branch = "main"
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "runner.toml"), []byte(`runner_type = "aws"
helm_driver = "configmap"
init_script_url = "https://example.com/init.sh"
`), 0o644))

	_, err := ParseDirWithSource(context.Background(), ParseConfig{
		Dirname:       dir,
		FileProcessor: func(_ string, obj map[string]any) map[string]any { return obj },
	})
	require.NoError(t, err)
}

const maxSourceArchiveTestFileBytes = 5<<20 + 1
