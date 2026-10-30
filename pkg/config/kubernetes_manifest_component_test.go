package config

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
)

func TestKubernetesManifestKustomizeValidation(t *testing.T) {
	tests := map[string]struct {
		options string
		wantErr string
	}{
		"omitted Helm":              {},
		"disabled Helm":             {options: "enable_helm = false"},
		"enabled Helm":              {options: "enable_helm = true", wantErr: "kustomize.enable_helm is not supported; omit it or set it to false and use a separate helm_chart component"},
		"explicit root restriction": {options: `load_restrictor = "rootOnly"`},
		"unrestricted loads":        {options: `load_restrictor = "none"`, wantErr: "kustomize.load_restrictor"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, repo := range []string{"public_repo", "connected_repo"} {
				t.Run(repo, func(t *testing.T) {
					var cfg KubernetesManifestComponentConfig
					require.NoError(t, toml.Unmarshal([]byte("namespace = \"default\"\n["+repo+"]\nrepo = \"acme/manifests\"\n[kustomize]\npath = \"overlays/production\"\n"+tc.options), &cfg))
					if tc.wantErr != "" {
						require.ErrorContains(t, cfg.Validate(), tc.wantErr)
						return
					}
					require.NoError(t, cfg.Validate())
				})
			}
		})
	}

	t.Run("inline manifest without Kustomize", func(t *testing.T) {
		cfg := KubernetesManifestComponentConfig{Manifest: "kind: ConfigMap", Namespace: "default"}
		require.NoError(t, cfg.Validate())
	})
}
