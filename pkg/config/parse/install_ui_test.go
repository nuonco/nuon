package parse

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func TestLoadInstallUIDefinitionOptional(t *testing.T) {
	require.NoError(t, LoadInstallUIDefinition(context.Background(), nil, "", ""))
	require.NoError(t, LoadInstallUIDefinition(context.Background(), &config.Install{}, "", ""))
	require.NoError(t, LoadInstallUIDefinition(context.Background(), &config.Install{
		StackOverrides: &config.InstallStackOverrides{},
	}, "", ""))
}

func TestLoadInstallUIDefinitionReadsFile(t *testing.T) {
	dir := t.TempDir()
	body := `{"handler":"Microsoft.Azure.CreateUIDef"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ui.json"), []byte(body), 0o644))

	cfg := &config.Install{
		StackOverrides: &config.InstallStackOverrides{UIDefinition: "./ui.json"},
	}
	require.NoError(t, LoadInstallUIDefinition(context.Background(), cfg, dir, dir))
	require.Equal(t, body, cfg.StackOverrides.UIDefinitionContents)
	require.Equal(t, "./ui.json", cfg.StackOverrides.UIDefinition)
}

func TestLoadInstallUIDefinitionResolvesFromInstallDir(t *testing.T) {
	root := t.TempDir()
	installs := filepath.Join(root, "installs")
	require.NoError(t, os.MkdirAll(installs, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "ui.json"), []byte("root"), 0o644))

	cfg := &config.Install{
		StackOverrides: &config.InstallStackOverrides{UIDefinition: "../ui.json"},
	}
	require.NoError(t, LoadInstallUIDefinition(context.Background(), cfg, root, installs))
	require.Equal(t, "root", cfg.StackOverrides.UIDefinitionContents)
}

func TestLoadInstallUIDefinitionRejectsEscapes(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, os.WriteFile(outside, []byte("secret"), 0o644))

	root := t.TempDir()
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "link.json")))

	for _, src := range []string{"file://" + outside, "../" + filepath.Base(filepath.Dir(outside)) + "/secret", "./link.json"} {
		cfg := &config.Install{
			StackOverrides: &config.InstallStackOverrides{UIDefinition: src},
		}
		require.Error(t, LoadInstallUIDefinition(context.Background(), cfg, root, root), src)
	}
}
