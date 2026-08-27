package config

import (
	"testing"

	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/stretchr/testify/require"
)

func TestNormalizeIntermediateConfigSortsComponentsAndManifests(t *testing.T) {
	cfg := &AppConfig{
		Components: ComponentList{
			{
				Name: "zulu",
				Type: KubernetesManifestComponentType,
				KubernetesManifest: &KubernetesManifestComponentConfig{
					Manifest: "  apiVersion: v1\r\nkind: ConfigMap  ",
					Kustomize: &KustomizeConfig{
						Path:    "overlays/prod",
						Patches: []string{"b.yaml", "a.yaml"},
					},
				},
			},
			{
				Name: "alpha",
				Type: KubernetesManifestComponentType,
				KubernetesManifest: &KubernetesManifestComponentConfig{
					Manifest: "kind: Namespace",
				},
			},
		},
	}

	NormalizeIntermediateConfig(cfg)

	require.Equal(t, "alpha", cfg.Components[0].Name)
	require.Equal(t, "zulu", cfg.Components[1].Name)
	require.Equal(t, "apiVersion: v1\nkind: ConfigMap", cfg.Components[1].KubernetesManifest.Manifest)
	require.Equal(t, []string{"a.yaml", "b.yaml"}, cfg.Components[1].KubernetesManifest.Kustomize.Patches)
}

func TestNormalizeIntermediateConfigIdenticalManifestsDoNotDiff(t *testing.T) {
	old := baseConfig()
	old.Components = ComponentList{
		{
			Name: "deploy",
			Type: KubernetesManifestComponentType,
			KubernetesManifest: &KubernetesManifestComponentConfig{
				Manifest:  "apiVersion: v1\nkind: ConfigMap",
				Namespace: "default",
			},
		},
	}
	newCfg := baseConfig()
	newCfg.Components = ComponentList{
		{
			Name: "deploy",
			Type: KubernetesManifestComponentType,
			KubernetesManifest: &KubernetesManifestComponentConfig{
				Manifest:  "  apiVersion: v1\r\nkind: ConfigMap  ",
				Namespace: "default",
			},
		},
	}

	NormalizeIntermediateConfig(old)
	NormalizeIntermediateConfig(newCfg)

	d := newCfg.Diff(old)
	comps := findChild(d, "components")
	require.NotNil(t, comps)
	deploy := findChild(comps, "component.deploy")
	require.NotNil(t, deploy)
	manifest := findChild(deploy, "manifest")
	require.NotNil(t, manifest)
	require.NotNil(t, manifest.Diff)
	require.Equal(t, diff.OpNoop, manifest.Diff.Op)
}
