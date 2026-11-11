package parse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse/dir"
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

func TestParseDirSkipBranchesIgnoresInvalidBranch(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "branches"), 0755))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "components"), 0755))
	files := map[string]string{
		"metadata.toml": `version = "v2"`,
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
		"branches/default.toml": `name = "default"

[[install_groups]]
name = "canary"
order = 1
`,
	}
	for name, contents := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0644))
	}
	fileProcessor := func(_ string, obj map[string]any) map[string]any { return obj }

	_, err := ParseDir(context.Background(), ParseConfig{
		Dirname:       dir,
		FileProcessor: fileProcessor,
	})
	require.ErrorContains(t, err, `branch "default": one install group must be default`)

	cfg, err := ParseDir(context.Background(), ParseConfig{
		Dirname:       dir,
		FileProcessor: fileProcessor,
		SkipBranches:  true,
	})
	require.NoError(t, err)
	require.Len(t, cfg.Branches, 1)
	require.Equal(t, "default", cfg.Branches[0].Name)
}

func TestParseDirSkipBranchesIgnoresUnparseableBranchFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "branches"), 0755))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "components"), 0755))
	files := map[string]string{
		"metadata.toml": `version = "v2"`,
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
		"branches/default.toml": `name = "default"

[[install_groups]]
name = "everyone"
order = 1
default = true
`,
		"branches/broken.toml": "this is not toml {{{",
	}
	for name, contents := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0644))
	}
	fileProcessor := func(_ string, obj map[string]any) map[string]any { return obj }

	_, err := ParseDir(context.Background(), ParseConfig{
		Dirname:       dir,
		FileProcessor: fileProcessor,
	})
	require.Error(t, err)

	cfg, err := ParseDir(context.Background(), ParseConfig{
		Dirname:       dir,
		FileProcessor: fileProcessor,
		SkipBranches:  true,
	})
	require.NoError(t, err)
	require.Len(t, cfg.Branches, 1)
	require.Equal(t, "default", cfg.Branches[0].Name)
}

func TestParseDirSkipInstallsIgnoresInstallFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "installs"), 0755))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "components"), 0755))
	files := map[string]string{
		"metadata.toml": `version = "v2"`,
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
		"installs.toml": `[connected_repo]
repo = "acme"
directory = "."
branch = "main"
`,
		"installs/foo.toml": "this is not toml {{{",
	}
	for name, contents := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0644))
	}
	fileProcessor := func(_ string, obj map[string]any) map[string]any { return obj }

	_, err := ParseDir(context.Background(), ParseConfig{
		Dirname:       dir,
		FileProcessor: fileProcessor,
	})
	require.Error(t, err)

	cfg, err := ParseDir(context.Background(), ParseConfig{
		Dirname:       dir,
		FileProcessor: fileProcessor,
		SkipInstalls:  true,
	})
	require.NoError(t, err)
	require.Empty(t, cfg.Installs)
	require.NotNil(t, cfg.InstallsConfig)
	require.NotNil(t, cfg.InstallsConfig.ConnectedRepo)
	require.Equal(t, "acme", cfg.InstallsConfig.ConnectedRepo.Repo)
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

func TestSourceFileRecorderBuildsMembersIndex(t *testing.T) {
	newSource := func() *sourceCapture {
		return &sourceCapture{archive: config.NewSourceArchive()}
	}

	t.Run("maps definition files to their members", func(t *testing.T) {
		source := newSource()
		rec := sourceFileRecorder(source)

		require.NoError(t, rec(dir.ParsedFile{
			Path:     "components/whoami.toml",
			Group:    "components",
			Contents: []byte("name = \"whoami\""),
			Value:    &struct{ Name string }{Name: "whoami"},
		}))
		require.NoError(t, rec(dir.ParsedFile{
			Path:     "actions/healthcheck.toml",
			Group:    "actions",
			Contents: []byte("name = \"healthcheck\""),
			Value:    &struct{ Name string }{Name: "healthcheck"},
		}))

		require.Equal(t, "components/whoami.toml", source.archive.Members["component:whoami"])
		require.Equal(t, "actions/healthcheck.toml", source.archive.Members["action:healthcheck"])
		require.Equal(t, "name = \"whoami\"", source.archive.Files["components/whoami.toml"])
		require.NoError(t, source.err)
	})

	t.Run("maps singleton group files to fixed members", func(t *testing.T) {
		source := newSource()
		rec := sourceFileRecorder(source)

		require.NoError(t, rec(dir.ParsedFile{Path: "metadata.toml", Group: "metadata", Value: &struct{ Name string }{}}))
		require.NoError(t, rec(dir.ParsedFile{Path: "inputs.toml", Group: "inputs", Value: &struct{ Name string }{}}))

		require.Equal(t, "metadata.toml", source.archive.Members["metadata:metadata"])
		require.Equal(t, "inputs.toml", source.archive.Members["input:inputs"])
	})

	t.Run("records an error when a definition has no name", func(t *testing.T) {
		source := newSource()
		rec := sourceFileRecorder(source)

		require.NoError(t, rec(dir.ParsedFile{
			Path:  "components/whoami.toml",
			Group: "components",
			Value: &struct{ Name string }{},
		}))

		require.ErrorContains(t, source.err, "has no name field")
	})

	t.Run("ignores files outside known groups", func(t *testing.T) {
		source := newSource()
		rec := sourceFileRecorder(source)

		require.NoError(t, rec(dir.ParsedFile{
			Path:  "nuon.toml",
			Group: "nuon",
			Value: &struct{ Name string }{Name: "nuon"},
		}))

		require.Empty(t, source.archive.Members)
		require.NoError(t, source.err)
	})
}

func TestLocalSourceFileRecorder(t *testing.T) {
	t.Run("is nil without a capture", func(t *testing.T) {
		require.Nil(t, localSourceFileRecorder(nil))
	})

	t.Run("records the root-relative path", func(t *testing.T) {
		source := &sourceCapture{archive: config.NewSourceArchive()}
		rec := localSourceFileRecorder(source)
		require.NotNil(t, rec)

		require.NoError(t, rec("values/base.yaml", []byte("replicas: 1")))

		require.Equal(t, "replicas: 1", source.archive.Files["values/base.yaml"])
		require.NoError(t, source.err)
	})
}

func TestSourceMemberIdentity(t *testing.T) {
	tests := []struct {
		group    string
		expected string
	}{
		{group: "components", expected: "component"},
		{group: "actions", expected: "action"},
		{group: "runbooks", expected: "runbook"},
		{group: "permissions", expected: "permission"},
		{group: "unknown", expected: ""},
	}
	for _, tt := range tests {
		t.Run(tt.group, func(t *testing.T) {
			kind, name := sourceMemberIdentity(tt.group, "components/x.toml", &struct{ Name string }{Name: "x"})
			require.Equal(t, tt.expected, kind)
			if tt.expected != "" {
				require.Equal(t, "x", name)
			} else {
				require.Empty(t, name)
			}
		})
	}

	t.Run("returns empty identity for a nil definition pointer", func(t *testing.T) {
		kind, name := sourceMemberIdentity("components", "components/x.toml", (*struct{ Name string })(nil))
		require.Empty(t, kind)
		require.Empty(t, name)
	})
}

func TestSourceMemberIdentityNamedIAMPolicyUsesFileStem(t *testing.T) {
	kind, name := sourceMemberIdentity("permissions/policies", "permissions/policies/logs.toml", &config.NamedIAMPolicy{
		Name: "install-{{.nuon.install.id}}-logs",
	})
	require.Equal(t, "permission_policy", kind)
	require.Equal(t, "logs", name)
}

func TestParseDirWithSourceComponentSymlinks(t *testing.T) {
	const component = `name = "acme-chart"
type = "helm_chart"
chart_name = "acme-chart"
[helm_repo]
repo_url = "https://example.com"
chart = "acme-chart"
version = "1.0.0"
`
	tests := map[string]struct {
		absolute bool
		outside  bool
		dir      bool
		nested   bool
	}{
		"relative file":              {},
		"relative directory":         {dir: true},
		"absolute in-repo file":      {absolute: true},
		"absolute in-repo directory": {absolute: true, dir: true},
		"nested absolute file":       {absolute: true, nested: true},
		"absolute outside file":      {absolute: true, outside: true},
		"relative outside file":      {outside: true},
		"relative outside directory": {outside: true, dir: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			repo := filepath.Join(root, "repo")
			configDir := filepath.Join(repo, "nuon")
			components := filepath.Join(configDir, "components")
			require.NoError(t, os.MkdirAll(components, 0o700))
			shared := filepath.Join(repo, "shared")
			if tc.outside {
				shared = filepath.Join(root, "outside")
			}
			require.NoError(t, os.MkdirAll(shared, 0o700))
			require.NoError(t, os.WriteFile(filepath.Join(shared, "chart.toml"), []byte(component), 0o600))
			files := map[string]string{
				"metadata.toml": `version = "v2"`,
				"sandbox.toml": `terraform_version = "1.11.3"
[public_repo]
repo = "acme/example-sandbox"
directory = "."
branch = "main"
`,
				"runner.toml": `runner_type = "aws"
helm_driver = "configmap"
init_script_url = "https://example.com/init.sh"
`,
			}
			for path, contents := range files {
				require.NoError(t, os.WriteFile(filepath.Join(configDir, path), []byte(contents), 0o600))
			}
			link := filepath.Join(components, "chart.toml")
			target := filepath.Join(shared, "chart.toml")
			if tc.dir {
				link = filepath.Join(components, "charts")
				target = shared
			}
			if tc.nested {
				require.NoError(t, os.Mkdir(filepath.Join(components, "nested"), 0o700))
				link = filepath.Join(components, "nested", "chart.toml")
			}
			if !tc.absolute {
				var err error
				target, err = filepath.Rel(filepath.Dir(link), target)
				require.NoError(t, err)
			}
			require.NoError(t, os.Symlink(target, link))
			result, err := ParseDirWithSource(context.Background(), ParseConfig{
				Dirname: configDir, SourceRoot: repo,
				FileProcessor: func(_ string, obj map[string]any) map[string]any { return obj },
			})
			if tc.absolute || tc.outside {
				require.Nil(t, result)
				require.ErrorContains(t, err, "unable to resolve config symlink")
				require.ErrorContains(t, err, "use a relative symlink whose target stays within the repository")
				linkPath, relErr := filepath.Rel(configDir, link)
				require.NoError(t, relErr)
				require.ErrorContains(t, err, linkPath)
				return
			}
			require.NoError(t, err)
			require.Len(t, result.Config.Components, 1)
			require.Equal(t, "acme-chart", result.Config.Components[0].Name)
			path := "components/chart.toml"
			if tc.dir {
				path = "components/charts/chart.toml"
			} else {
				require.NotEmpty(t, result.Config.Components[0].Checksum)
			}
			require.Equal(t, component, result.Source.Files[path])
			require.Equal(t, path, result.Source.Members["component:acme-chart"])
		})
	}
}

func TestParseDirWithSourceConfinesWorkerImports(t *testing.T) {
	const canary = "outside-worker-canary"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, canary)
	}))
	t.Cleanup(server.Close)
	tests := map[string]struct {
		reference       func(string) string
		wantError       bool
		configSubdir    string
		repositoryRoot  bool
		tomlSymlink     bool
		gcpModule       bool
		rawValues       bool
		componentSource bool
	}{
		"repository file":              {reference: func(string) string { return "./inside.txt" }},
		"authored hidden file":         {reference: func(string) string { return "./.config/inside.txt" }},
		"Git metadata":                 {reference: func(string) string { return "./.git/config" }, wantError: true},
		"Git metadata file alias":      {reference: func(string) string { return "./metadata-alias" }, wantError: true},
		"Git metadata directory alias": {reference: func(string) string { return "./metadata-dir/config" }, wantError: true},
		"repository sibling": {
			reference: func(string) string { return "../shared.txt" }, configSubdir: "config", repositoryRoot: true,
		},
		"sibling without repository root": {
			reference: func(string) string { return "../shared.txt" }, configSubdir: "config", wantError: true,
		},
		"outside explicit repository root": {
			reference: func(string) string { return "../../outside.txt" }, configSubdir: "config", repositoryRoot: true, wantError: true,
		},
		"outside absolute":                    {reference: func(root string) string { return "file://" + filepath.Join(root, "outside.txt") }, wantError: true},
		"outside traversal":                   {reference: func(string) string { return "../outside.txt" }, wantError: true},
		"outside encoded path":                {reference: func(root string) string { return "file://" + filepath.Join(root, "repo") + "/%2e%2e/outside.txt" }, wantError: true},
		"outside symlink":                     {reference: func(string) string { return "./escape.txt" }, wantError: true},
		"outside TOML symlink":                {reference: func(string) string { return "./inside.txt" }, tomlSymlink: true, wantError: true},
		"internal HTTP":                       {reference: func(string) string { return server.URL + "/canary" }, wantError: true},
		"GCP module reference is not fetched": {reference: func(string) string { return "git::https://example.com/acme/modules.git//gcp/modules/network" }, gcpModule: true},
		"raw values file":                     {reference: func(string) string { return "./inside.txt" }, rawValues: true},
		"raw values sibling":                  {reference: func(string) string { return "../shared.txt" }, configSubdir: "config", repositoryRoot: true, rawValues: true},
		"raw values outside":                  {reference: func(root string) string { return filepath.Join(root, "outside.txt") }, rawValues: true, wantError: true},
		"raw values symlink":                  {reference: func(string) string { return "./escape.txt" }, rawValues: true, wantError: true},
		"component source inside":             {reference: func(string) string { return "./source-component.toml" }, componentSource: true},
		"component source outside":            {reference: func(root string) string { return filepath.Join(root, "outside-component.toml") }, componentSource: true, wantError: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			repo := filepath.Join(root, "repo")
			configDir := filepath.Join(repo, tc.configSubdir)
			require.NoError(t, os.MkdirAll(filepath.Join(configDir, "components"), 0o700))
			require.NoError(t, os.MkdirAll(filepath.Join(configDir, ".git"), 0o700))
			require.NoError(t, os.MkdirAll(filepath.Join(configDir, ".config"), 0o700))
			require.NoError(t, os.WriteFile(filepath.Join(configDir, ".git/config"), []byte(canary), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(configDir, ".config/inside.txt"), []byte("repository contents"), 0o600))
			require.NoError(t, os.Symlink(".git/config", filepath.Join(configDir, "metadata-alias")))
			require.NoError(t, os.Symlink(".git", filepath.Join(configDir, "metadata-dir")))
			require.NoError(t, os.WriteFile(filepath.Join(repo, "shared.txt"), []byte("repository contents"), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(root, "outside.txt"), []byte(canary), 0o600))
			require.NoError(t, os.Symlink(filepath.Join(root, "outside.txt"), filepath.Join(configDir, "escape.txt")))
			reference := tc.reference(root)
			if tc.rawValues || tc.componentSource {
				reference = "./inside.txt"
			}
			files := map[string]string{
				"metadata.toml": `version = "v2"`,
				"sandbox.toml": `terraform_version = "1.11.3"
[public_repo]
repo = "acme/example-sandbox"
directory = "."
branch = "main"
`,
				"runner.toml": `runner_type = "aws"
helm_driver = "configmap"
init_script_url = "https://example.com/init.sh"
`,
				"inside.txt": "repository contents",
				"stack.toml": fmt.Sprintf(`type = "aws-cloudformation"
name = "acme-stack"
vpc_nested_template_url = "https://s3.amazonaws.com/acme/vpc.yaml"
runner_nested_template_url = "https://s3.amazonaws.com/acme/runner.yaml"
[[custom_nested_stacks]]
name = "custom-stack"
index = 0
template_url = %q
`, reference),
			}
			if tc.gcpModule {
				files["stack.toml"] = strings.Replace(files["stack.toml"], "aws-cloudformation", "gcp-terraform", 1)
			}
			if tc.rawValues || tc.componentSource {
				component := `name = "acme-chart"
type = "helm_chart"
chart_name = "acme-chart"
[helm_repo]
repo_url = "https://example.com"
chart = "acme-chart"
version = "1.0.0"
`
				if tc.rawValues {
					files["components/chart.toml"] = component + fmt.Sprintf("[[values_file]]\npath = %q\n", tc.reference(root))
				} else {
					files["components/chart.toml"] = fmt.Sprintf("source = %q\n", tc.reference(root))
					files["source-component.toml"] = component
					require.NoError(t, os.WriteFile(filepath.Join(root, "outside-component.toml"), []byte(component), 0o600))
				}
			}
			for path, contents := range files {
				require.NoError(t, os.WriteFile(filepath.Join(configDir, path), []byte(contents), 0o600))
			}
			if tc.tomlSymlink {
				outside := filepath.Join(root, "outside-stack.toml")
				require.NoError(t, os.WriteFile(outside, []byte(files["stack.toml"]), 0o600))
				require.NoError(t, os.Remove(filepath.Join(configDir, "stack.toml")))
				require.NoError(t, os.Symlink(outside, filepath.Join(configDir, "stack.toml")))
			}
			parseCfg := ParseConfig{
				Dirname: configDir, SkipBranches: true, SkipInstalls: true,
				FileProcessor: func(_ string, obj map[string]any) map[string]any { return obj },
			}
			if tc.repositoryRoot {
				parseCfg.SourceRoot = repo
			}
			result, err := ParseDirWithSource(context.Background(), parseCfg)
			if tc.wantError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tc.gcpModule {
					require.Equal(t, tc.reference(root), result.Config.Stack.CustomNestedStacks[0].TemplateURL)
					require.Empty(t, result.Config.Stack.CustomNestedStacks[0].Contents)
				} else {
					require.Equal(t, "repository contents", result.Config.Stack.CustomNestedStacks[0].Contents)
				}
				if tc.rawValues {
					require.Equal(t, "repository contents", result.Config.Components[0].HelmChart.ValuesFiles[0].Contents)
				}
				if tc.componentSource {
					require.Equal(t, "acme-chart", result.Config.Components[0].HelmChart.ChartName)
				}
			}
		})
	}
}
