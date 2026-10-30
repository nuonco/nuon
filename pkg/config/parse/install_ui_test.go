package parse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse/get"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

func TestLoadInstallUIDefinitionOptional(t *testing.T) {
	require.NoError(t, LoadInstallUIDefinition(context.Background(), nil, nil))
	require.NoError(t, LoadInstallUIDefinition(context.Background(), &config.Install{}, nil))
	require.NoError(t, LoadInstallUIDefinition(context.Background(), &config.Install{
		StackOverrides: &config.InstallStackOverrides{},
	}, nil))
}

func TestLoadInstallUIDefinitionReadsFile(t *testing.T) {
	dir := t.TempDir()
	body := `{"handler":"Microsoft.Azure.CreateUIDef"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ui.json"), []byte(body), 0o644))

	cfg := &config.Install{
		StackOverrides: &config.InstallStackOverrides{UIDefinition: "./ui.json"},
	}
	var recorded []string
	opts := &get.Options{
		RootDir: dir, FieldTimeout: 2 * time.Second,
		OnLocalFile: func(name string, contents []byte) error {
			recorded = append(recorded, name+":"+string(contents))
			return nil
		},
	}
	require.NoError(t, LoadInstallUIDefinition(context.Background(), cfg, opts))
	require.Equal(t, body, cfg.StackOverrides.UIDefinitionContents)
	require.Equal(t, "./ui.json", cfg.StackOverrides.UIDefinition)
	require.Equal(t, []string{"ui.json:" + body}, recorded)
	require.Equal(t, 2*time.Second, opts.FieldTimeout)
}

func TestLoadInstallUIDefinitionUsesRepositorySource(t *testing.T) {
	for _, backend := range []string{"os", "memory"} {
		t.Run(backend, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "repo")
			require.NoError(t, os.MkdirAll(filepath.Join(root, "installs"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(root, "ui.json"), []byte(`{"source":"host canary"}`), 0o644))
			var fs afero.Fs = afero.NewOsFs()
			if backend == "memory" {
				fs = afero.NewMemMapFs()
			}
			body := `{"source":"repository"}`
			require.NoError(t, fs.MkdirAll(filepath.Join(root, "installs"), 0o755))
			require.NoError(t, afero.WriteFile(fs, filepath.Join(root, "ui.json"), []byte(body), 0o644))
			source, err := sourcefs.New(fs, root)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, source.Close()) })
			if backend == "os" {
				require.NoError(t, os.Rename(root, root+"-pinned"))
				require.NoError(t, os.MkdirAll(filepath.Join(root, "installs"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(root, "ui.json"), []byte(`{"source":"replacement canary"}`), 0o644))
			}

			cfg := &config.Install{StackOverrides: &config.InstallStackOverrides{UIDefinition: "../ui.json"}}
			opts := &get.Options{RootDir: filepath.Join(root, "installs"), Source: source}
			require.NoError(t, LoadInstallUIDefinition(context.Background(), cfg, opts))
			require.Equal(t, body, cfg.StackOverrides.UIDefinitionContents)
			require.Equal(t, "../ui.json", cfg.StackOverrides.UIDefinition)
			require.Zero(t, opts.FieldTimeout)
			contents, err := source.ReadFile("ui.json")
			require.NoError(t, err)
			require.Equal(t, body, string(contents))
		})
	}
}

func TestLoadInstallUIDefinitionRejectsRepositoryEscape(t *testing.T) {
	root := filepath.Join(t.TempDir(), "repo")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "installs"), 0o755))
	outside := filepath.Join(filepath.Dir(root), "outside.json")
	require.NoError(t, os.WriteFile(outside, []byte("outside canary"), 0o644))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "installs/link.json")))
	source, err := sourcefs.New(afero.NewOsFs(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	for _, reference := range []string{"../../outside.json", "./link.json", "file://" + outside, "file::file://" + outside} {
		t.Run(reference, func(t *testing.T) {
			cfg := &config.Install{StackOverrides: &config.InstallStackOverrides{UIDefinition: reference}}
			err := LoadInstallUIDefinition(context.Background(), cfg, &get.Options{RootDir: filepath.Join(root, "installs"), Source: source})
			require.ErrorContains(t, err, "ui_definition:")
			require.NotEqual(t, "outside canary", cfg.StackOverrides.UIDefinitionContents)
		})
	}
}

func TestLoadInstallUIDefinitionDefaultsTimeoutWithoutMutatingOptions(t *testing.T) {
	body := `{"handler":"Microsoft.Azure.CreateUIDef"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()
	opts := &get.Options{RootDir: t.TempDir()}
	cfg := &config.Install{StackOverrides: &config.InstallStackOverrides{UIDefinition: server.URL + "/ui.json"}}
	require.NoError(t, LoadInstallUIDefinition(context.Background(), cfg, opts))
	require.Equal(t, body, cfg.StackOverrides.UIDefinitionContents)
	require.Zero(t, opts.FieldTimeout)
}
