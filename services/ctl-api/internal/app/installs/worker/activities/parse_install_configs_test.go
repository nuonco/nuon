package activities

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/sourcefs"
	"github.com/nuonco/nuon/pkg/urlpolicy"
)

const policyRejection = "is not allowed by this control plane"

func installSourceWithUIDefinition(t *testing.T, uiDefinition string) *sourcefs.FS {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "installs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "installs", "ui.json"), []byte(`{"handler":"Microsoft.Azure.CreateUIDef"}`), 0o644))
	contents := "name = \"acme\"\n\n[stack_overrides]\nui_definition = \"" + uiDefinition + "\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "installs", "acme.toml"), []byte(contents), 0o644))
	source, err := sourcefs.New(afero.NewOsFs(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	return source
}

func stackTemplatesPolicy(t *testing.T, mode urlpolicy.Mode) (*urlpolicy.Policy, *urlpolicy.Violations) {
	t.Helper()
	policy, err := urlpolicy.New(urlpolicy.Config{
		Mode: mode,
		Categories: map[urlpolicy.Category]urlpolicy.CategoryConfig{
			urlpolicy.CategoryStackTemplates: {Entries: []string{"https://templates.acme.example.com"}},
		},
	})
	require.NoError(t, err)
	warned := &urlpolicy.Violations{}
	policy.OnWarn(func(violations urlpolicy.Violations) { *warned = append(*warned, violations...) })
	return policy, warned
}

// The off-list URL is a closed loopback port, so a fetch that gets past the
// policy fails on its own without leaving the machine.
func TestParseInstallTOMLFileChecksRemoteUIDefinition(t *testing.T) {
	const offList = "https://127.0.0.1:1/ui.json"
	installFile := filepath.Join("installs", "acme.toml")

	t.Run("enforce rejects an off-list url", func(t *testing.T) {
		policy, warned := stackTemplatesPolicy(t, urlpolicy.ModeEnforce)
		_, err := parseInstallTOMLFile(context.Background(), installSourceWithUIDefinition(t, offList), installFile, policy)
		require.ErrorContains(t, err, policyRejection)
		require.ErrorContains(t, err, urlpolicy.CategoryStackTemplates.ConfigKey())
		require.Empty(t, *warned)
	})

	t.Run("warn reports an off-list url and lets the fetch go ahead", func(t *testing.T) {
		policy, warned := stackTemplatesPolicy(t, urlpolicy.ModeWarn)
		_, err := parseInstallTOMLFile(context.Background(), installSourceWithUIDefinition(t, offList), installFile, policy)
		require.Error(t, err)
		require.NotContains(t, err.Error(), policyRejection)
		require.Len(t, *warned, 1)
		require.Equal(t, urlpolicy.CategoryStackTemplates, (*warned)[0].Category)
		require.Equal(t, offList, (*warned)[0].Value)
	})

	t.Run("an unrestricted policy does not check the url", func(t *testing.T) {
		_, err := parseInstallTOMLFile(context.Background(), installSourceWithUIDefinition(t, offList), installFile, nil)
		require.Error(t, err)
		require.NotContains(t, err.Error(), policyRejection)
	})

	t.Run("a local path is not checked", func(t *testing.T) {
		policy, warned := stackTemplatesPolicy(t, urlpolicy.ModeEnforce)
		cfg, err := parseInstallTOMLFile(context.Background(), installSourceWithUIDefinition(t, "./ui.json"), installFile, policy)
		require.NoError(t, err)
		require.JSONEq(t, `{"handler":"Microsoft.Azure.CreateUIDef"}`, cfg.StackOverrides.UIDefinitionContents)
		require.Empty(t, *warned)
	})
}
