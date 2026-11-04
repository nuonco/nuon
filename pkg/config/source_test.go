package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

type configImportPath struct {
	name   string
	path   string
	denied bool
}

func configImportSource(t *testing.T, contents string) (*sourcefs.FS, string, []configImportPath) {
	t.Helper()
	parent := t.TempDir()
	root := filepath.Join(parent, "repo")
	configDir := filepath.Join(root, "apps", "app")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(root, "shared"), 0755))
	inside := filepath.Join(root, "shared", "import.toml")
	outside := filepath.Join(parent, "import.toml")
	require.NoError(t, os.WriteFile(inside, []byte(contents), 0644))
	require.NoError(t, os.WriteFile(outside, []byte(contents), 0644))
	require.NoError(t, os.Symlink("../../shared/import.toml", filepath.Join(configDir, "allowed.toml")))
	require.NoError(t, os.Symlink(outside, filepath.Join(configDir, "escape.toml")))
	require.NoError(t, os.Symlink(parent, filepath.Join(configDir, "escape-dir")))
	source, err := sourcefs.New(afero.NewOsFs(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	return source, configDir, []configImportPath{
		{name: "shared", path: "../../shared/import.toml"},
		{name: "absolute-inside", path: inside},
		{name: "symlink-inside", path: "allowed.toml"},
		{name: "absolute-outside", path: outside, denied: true},
		{name: "traversal", path: "../../../import.toml", denied: true},
		{name: "symlink-outside", path: "escape.toml", denied: true},
		{name: "directory-symlink-outside", path: "escape-dir/import.toml", denied: true},
	}
}

func TestAppConfigSourceImports(t *testing.T) {
	for _, route := range []struct {
		name      string
		contents  string
		configure func(*config.AppConfig, string)
		loaded    func(*config.AppConfig) string
		want      string
	}{
		{
			name:     "sandbox",
			contents: "terraform_version = '1.5.0'\n[public_repo]\nrepo = 'https://example.com/acme/repo'\n",
			configure: func(app *config.AppConfig, path string) {
				app.Sandbox.Source = path
			},
			loaded: func(app *config.AppConfig) string { return app.Sandbox.TerraformVersion },
			want:   "1.5.0",
		},
		{
			name:     "runner",
			contents: "instance_type = 'trusted'\n",
			configure: func(app *config.AppConfig, path string) {
				app.Runner.Source = path
			},
			loaded: func(app *config.AppConfig) string { return app.Runner.InstanceType },
			want:   "trusted",
		},
		{
			name:     "inputs-source",
			contents: "[[input]]\nname = 'trusted'\ntype = 'string'\n",
			configure: func(app *config.AppConfig, path string) {
				app.Inputs = &config.AppInputConfig{Source: path}
			},
			loaded: func(app *config.AppConfig) string { return app.Inputs.Inputs[0].Name },
			want:   "trusted",
		},
		{
			name:     "inputs-sources",
			contents: "[[input]]\nname = 'trusted'\ntype = 'string'\n",
			configure: func(app *config.AppConfig, path string) {
				app.Inputs = &config.AppInputConfig{Sources: []string{path}}
			},
			loaded: func(app *config.AppConfig) string { return app.Inputs.Inputs[0].Name },
			want:   "trusted",
		},
		{
			name:     "helm-path",
			contents: "replicas: 2\n",
			configure: func(app *config.AppConfig, path string) {
				app.Components = config.ComponentList{{
					Name: "chart", Type: config.HelmChartComponentType,
					HelmChart: &config.HelmChartComponentConfig{ValuesFiles: []config.HelmValuesFile{{Path: path}}},
				}}
			},
			loaded: func(app *config.AppConfig) string { return app.Components[0].HelmChart.ValuesFiles[0].Contents },
			want:   "replicas: 2\n",
		},
		{
			name:     "helm-source",
			contents: "replicas: 2\n",
			configure: func(app *config.AppConfig, path string) {
				app.Components = config.ComponentList{{
					Name: "chart", Type: config.HelmChartComponentType,
					HelmChart: &config.HelmChartComponentConfig{ValuesFiles: []config.HelmValuesFile{{Source: path}}},
				}}
			},
			loaded: func(app *config.AppConfig) string { return app.Components[0].HelmChart.ValuesFiles[0].Contents },
			want:   "replicas: 2\n",
		},
	} {
		t.Run(route.name, func(t *testing.T) {
			source, configDir, paths := configImportSource(t, route.contents)
			for _, path := range paths {
				t.Run(path.name, func(t *testing.T) {
					app := &config.AppConfig{Sandbox: &config.AppSandboxConfig{}, Runner: &config.AppRunnerConfig{}}
					route.configure(app, path.path)
					err := app.Parse(config.WithRootDir(configDir), config.WithSource(source))
					if path.denied {
						require.ErrorContains(t, err, "escapes repository")
						return
					}
					require.NoError(t, err)
					require.Equal(t, route.want, route.loaded(app))
				})
			}
			t.Run("legacy-nil-source", func(t *testing.T) {
				app := &config.AppConfig{Sandbox: &config.AppSandboxConfig{}, Runner: &config.AppRunnerConfig{}}
				route.configure(app, paths[3].path)
				require.NoError(t, app.Parse(config.WithRootDir(configDir), config.WithSource(nil)))
				require.Equal(t, route.want, route.loaded(app))
			})
			byts, err := source.ReadFile(filepath.Join(configDir, "allowed.toml"))
			require.NoError(t, err)
			require.Equal(t, route.contents, string(byts))
		})
	}
}

func TestDecoderConfigSourceImports(t *testing.T) {
	source, configDir, paths := configImportSource(t, "type = 'helm_chart'\nname = 'imported'\nchart_name = 'trusted'\n")
	for _, path := range paths {
		t.Run(path.name, func(t *testing.T) {
			var component config.Component
			opts := config.DecoderConfig(config.WithRootDir(configDir), config.WithSource(source))
			opts.Result = &component
			decoder, err := mapstructure.NewDecoder(opts)
			require.NoError(t, err)
			err = decoder.Decode(map[string]any{"source": path.path})
			if path.denied {
				require.ErrorContains(t, err, "escapes repository")
				require.Nil(t, component.HelmChart)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "imported", component.Name)
			require.NotNil(t, component.HelmChart)
			require.Equal(t, "trusted", component.HelmChart.ChartName)
		})
	}
}

func TestAppConfigPolicySourceUsesRepository(t *testing.T) {
	source, configDir, paths := configImportSource(t, "\n\n[[policy]]\ncontents = 'inline'\n")
	for _, path := range paths {
		t.Run(path.name, func(t *testing.T) {
			app := &config.AppConfig{
				Sandbox: &config.AppSandboxConfig{}, Runner: &config.AppRunnerConfig{},
				Policies: &config.PoliciesConfig{SourceFile: path.path, Policies: []config.AppPolicy{{Contents: "inline"}}},
			}
			require.NoError(t, app.Parse(config.WithRootDir(configDir), config.WithSource(source)))
			if path.denied {
				require.Zero(t, app.Policies.Policies[0].SourceLine)
			} else {
				require.Equal(t, 3, app.Policies.Policies[0].SourceLine)
			}
		})
	}
}
