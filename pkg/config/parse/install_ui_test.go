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
	require.NoError(t, LoadInstallUIDefinition(context.Background(), nil, ""))
	require.NoError(t, LoadInstallUIDefinition(context.Background(), &config.Install{}, ""))
	require.NoError(t, LoadInstallUIDefinition(context.Background(), &config.Install{
		StackOverrides: &config.InstallStackOverrides{},
	}, ""))
}

func TestLoadInstallUIDefinitionReadsFile(t *testing.T) {
	dir := t.TempDir()
	body := `{"handler":"Microsoft.Azure.CreateUIDef"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ui.json"), []byte(body), 0o644))

	cfg := &config.Install{
		StackOverrides: &config.InstallStackOverrides{UIDefinition: "./ui.json"},
	}
	require.NoError(t, LoadInstallUIDefinition(context.Background(), cfg, dir))
	require.Equal(t, body, cfg.StackOverrides.UIDefinitionContents)
	require.Equal(t, "./ui.json", cfg.StackOverrides.UIDefinition)
}
