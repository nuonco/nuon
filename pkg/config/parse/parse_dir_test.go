package parse

import (
	"context"
	"os"
	"path/filepath"
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

func TestParseDirResolvesHelmValuesFromConfigRoot(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "components"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "values"), 0755))

	files := map[string]string{
		"metadata.toml": `version = "v2"`,
		"sandbox.toml": `
terraform_version = "1.11.3"
[public_repo]
repo = "nuonco/aws-eks-sandbox"
directory = "."
branch = "main"
`,
		"runner.toml": `
runner_type = "aws"
helm_driver = "configmap"
init_script_url = "https://example.com/init.sh"
`,
		"components/chart.toml": `
name = "chart"
type = "helm_chart"
chart_name = "chart"
[helm_repo]
repo_url = "https://example.com"
chart = "chart"
version = "1.0.0"
[[values_file]]
source = "values/chart.yaml"
[[values_file]]
path = "./values/path.yaml"
`,
		"values/chart.yaml": "replicaCount: 2\n",
		"values/path.yaml":  "replicaCount: 4\n",
	}
	for name, contents := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0644))
	}

	t.Chdir(t.TempDir())
	cfg, err := ParseDir(context.Background(), ParseConfig{
		Dirname:       dir,
		FileProcessor: func(_ string, obj map[string]any) map[string]any { return obj },
	})
	require.NoError(t, err)
	require.Equal(t, "replicaCount: 2\n", cfg.Components[0].HelmChart.ValuesFiles[0].Contents)
	require.Equal(t, "replicaCount: 4\n", cfg.Components[0].HelmChart.ValuesFiles[1].Contents)
}

func TestParseDirGetsHelmValuesContentsFromComponentDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "components", "values"), 0755))

	files := map[string]string{
		"metadata.toml": `version = "v2"`,
		"sandbox.toml": `
terraform_version = "1.11.3"
[public_repo]
repo = "nuonco/aws-eks-sandbox"
directory = "."
branch = "main"
`,
		"runner.toml": `
runner_type = "aws"
helm_driver = "configmap"
init_script_url = "https://example.com/init.sh"
`,
		"components/chart.toml": `
name = "chart"
type = "helm_chart"
chart_name = "chart"
[helm_repo]
repo_url = "https://example.com"
chart = "chart"
version = "1.0.0"
[[values_file]]
contents = "./values/chart.yaml"
`,
		"components/values/chart.yaml": "replicaCount: 3\n",
	}
	for name, contents := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0644))
	}

	cfg, err := ParseDir(context.Background(), ParseConfig{
		Dirname:       dir,
		FileProcessor: func(_ string, obj map[string]any) map[string]any { return obj },
	})
	require.NoError(t, err)
	require.Equal(t, "replicaCount: 3\n", cfg.Components[0].HelmChart.ValuesFiles[0].Contents)
}
