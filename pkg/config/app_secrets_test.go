package config

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
)

// decodeAppSecret decodes a raw map (as produced by the TOML decoder) into an AppSecret using the same decoder config
// the real parse path uses.
func decodeAppSecret(t *testing.T, raw map[string]interface{}) *AppSecret {
	t.Helper()

	var out AppSecret
	cfg := DecoderConfig()
	cfg.Result = &out
	dec, err := mapstructure.NewDecoder(cfg)
	require.NoError(t, err)
	require.NoError(t, dec.Decode(raw))
	return &out
}

func TestKubernetesSyncTarget_NamespaceDecoding(t *testing.T) {
	secret := decodeAppSecret(t, map[string]interface{}{
		"name":        "datadog-api-key",
		"description": "DataDog API Key",
		"kubernetes_sync_targets": []interface{}{
			map[string]interface{}{
				"namespaces": []interface{}{"cloudprem", "datadog"},
				"name":       "datadog-api-key",
				"key":        "api-key",
			},
		},
	})

	require.Len(t, secret.KubernetesSyncTargets, 1)
	require.Equal(t, []string{"cloudprem", "datadog"}, secret.KubernetesSyncTargets[0].Namespaces)
	require.Equal(t, "datadog-api-key", secret.KubernetesSyncTargets[0].Name)
	require.Equal(t, "api-key", secret.KubernetesSyncTargets[0].Key)
}

func TestAppSecret_KubernetesSyncEnabled(t *testing.T) {
	require.False(t, (&AppSecret{}).KubernetesSyncEnabled())
	require.False(t, (&AppSecret{KubernetesSync: generics.ToPtr(false)}).KubernetesSyncEnabled())
	require.True(t, (&AppSecret{KubernetesSync: generics.ToPtr(true)}).KubernetesSyncEnabled())
	require.True(t, (&AppSecret{
		KubernetesSyncTargets: []*KubernetesSyncTarget{{Namespaces: []string{"ns"}, Name: "n", Key: "k"}},
	}).KubernetesSyncEnabled())
	// targets present + explicit false -> still enabled (targets win)
	require.True(t, (&AppSecret{
		KubernetesSync:        generics.ToPtr(false),
		KubernetesSyncTargets: []*KubernetesSyncTarget{{Namespaces: []string{"ns"}, Name: "n", Key: "k"}},
	}).KubernetesSyncEnabled())
}

func TestAppSecret_Validate_Targets(t *testing.T) {
	base := func() *AppSecret {
		return &AppSecret{Name: "s", Description: "d"}
	}

	t.Run("valid target", func(t *testing.T) {
		s := base()
		s.KubernetesSyncTargets = []*KubernetesSyncTarget{{Namespaces: []string{"ns"}, Name: "n", Key: "k"}}
		require.NoError(t, s.Validate())
	})

	t.Run("missing namespace", func(t *testing.T) {
		s := base()
		s.KubernetesSyncTargets = []*KubernetesSyncTarget{{Name: "n", Key: "k"}}
		require.Error(t, s.Validate())
	})

	t.Run("missing key", func(t *testing.T) {
		s := base()
		s.KubernetesSyncTargets = []*KubernetesSyncTarget{{Namespaces: []string{"ns"}, Name: "n"}}
		require.Error(t, s.Validate())
	})

	t.Run("legacy single-target still requires name and namespace", func(t *testing.T) {
		s := base()
		s.KubernetesSync = generics.ToPtr(true)
		require.Error(t, s.Validate())
	})

	t.Run("legacy requirement skipped when targets present", func(t *testing.T) {
		s := base()
		s.KubernetesSync = generics.ToPtr(true) // explicitly on, but using targets
		s.KubernetesSyncTargets = []*KubernetesSyncTarget{{Namespaces: []string{"ns"}, Name: "n", Key: "k"}}
		require.NoError(t, s.Validate())
	})

	t.Run("invalid namespace fails hostname_rfc1123", func(t *testing.T) {
		s := base()
		s.KubernetesSyncTargets = []*KubernetesSyncTarget{{Namespaces: []string{"bad_name_"}, Name: "whoami", Key: "bar"}}
		require.Error(t, s.Validate())
	})

	t.Run("invalid target name fails hostname_rfc1123", func(t *testing.T) {
		s := base()
		s.KubernetesSyncTargets = []*KubernetesSyncTarget{{Namespaces: []string{"ok"}, Name: "Bad_Name", Key: "bar"}}
		require.Error(t, s.Validate())
	})

	t.Run("templated namespace and name are allowed (validated after rendering)", func(t *testing.T) {
		s := base()
		s.KubernetesSyncTargets = []*KubernetesSyncTarget{{
			Namespaces: []string{"{{.nuon.install.id}}-ns"},
			Name:       "{{.nuon.install.id}}-secret",
			Key:        "bar",
		}}
		require.NoError(t, s.Validate())
	})
}

func TestSecretsConfig_Validate_Collision(t *testing.T) {
	t.Run("same secret/key from two secrets is rejected", func(t *testing.T) {
		cfg := &SecretsConfig{Secrets: []*AppSecret{
			{Name: "a", Description: "d", KubernetesSyncTargets: []*KubernetesSyncTarget{
				{Namespaces: []string{"datadog"}, Name: "datadog", Key: "api-key"},
			}},
			{Name: "b", Description: "d", KubernetesSyncTargets: []*KubernetesSyncTarget{
				{Namespaces: []string{"datadog"}, Name: "datadog", Key: "api-key"},
			}},
		}}
		require.Error(t, cfg.Validate())
	})

	t.Run("same secret different keys is allowed", func(t *testing.T) {
		cfg := &SecretsConfig{Secrets: []*AppSecret{
			{Name: "a", Description: "d", KubernetesSyncTargets: []*KubernetesSyncTarget{
				{Namespaces: []string{"datadog"}, Name: "datadog", Key: "api-key"},
			}},
			{Name: "b", Description: "d", KubernetesSyncTargets: []*KubernetesSyncTarget{
				{Namespaces: []string{"datadog"}, Name: "datadog", Key: "app-key"},
			}},
		}}
		require.NoError(t, cfg.Validate())
	})

	t.Run("collision across namespaces in a single target", func(t *testing.T) {
		cfg := &SecretsConfig{Secrets: []*AppSecret{
			{Name: "a", Description: "d", KubernetesSyncTargets: []*KubernetesSyncTarget{
				{Namespaces: []string{"cloudprem", "datadog"}, Name: "datadog-api-key", Key: "api-key"},
			}},
			{Name: "b", Description: "d", KubernetesSyncTargets: []*KubernetesSyncTarget{
				{Namespaces: []string{"datadog"}, Name: "datadog-api-key", Key: "api-key"},
			}},
		}}
		require.Error(t, cfg.Validate())
	})
}

func TestSecretsConfig_Validate_ExplicitDisableWarning(t *testing.T) {
	t.Run("explicit kubernetes_sync=false with targets warns but does not error", func(t *testing.T) {
		cfg := &SecretsConfig{Secrets: []*AppSecret{
			{Name: "a", Description: "d", KubernetesSync: generics.ToPtr(false), KubernetesSyncTargets: []*KubernetesSyncTarget{
				{Namespaces: []string{"datadog"}, Name: "datadog", Key: "api-key"},
			}},
		}}
		err := cfg.Validate()
		require.Error(t, err)
		require.True(t, IsWarningErr(err))
	})

	t.Run("omitted kubernetes_sync with targets does not warn", func(t *testing.T) {
		cfg := &SecretsConfig{Secrets: []*AppSecret{
			{Name: "a", Description: "d", KubernetesSyncTargets: []*KubernetesSyncTarget{
				{Namespaces: []string{"datadog"}, Name: "datadog", Key: "api-key"},
			}},
		}}
		require.NoError(t, cfg.Validate())
	})

	t.Run("hard error takes precedence over warning", func(t *testing.T) {
		cfg := &SecretsConfig{Secrets: []*AppSecret{
			{Name: "a", Description: "d", KubernetesSync: generics.ToPtr(false), KubernetesSyncTargets: []*KubernetesSyncTarget{
				{Namespaces: []string{"bad_name_"}, Name: "datadog", Key: "api-key"},
			}},
		}}
		err := cfg.Validate()
		require.Error(t, err)
		require.False(t, IsWarningErr(err))
	})
}
