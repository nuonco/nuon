package config

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/config/diff"
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

func TestNormalizeIntermediateConfig_RoundTripProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		names := drawDistinctStrings(t, hegel.FromRegex(`[a-z]{1,8}`, true), 1, 6)
		patches := drawDistinctStrings(t, hegel.FromRegex(`[a-z]{1,8}`, true), 0, 5)
		large := json.Number(strconv.FormatInt(hegel.Draw(t, hegel.Integers[int64](1<<53+1, 1<<60)), 10))
		cfg := &AppConfig{Inputs: &AppInputConfig{Inputs: []AppInput{{Name: "capacity", Type: "number", Default: large}}}}
		for _, name := range drawPermutation(t, names) {
			cfg.Components = append(cfg.Components, &Component{Name: name, Type: KubernetesManifestComponentType, KubernetesManifest: &KubernetesManifestComponentConfig{Manifest: " \r\nkind: ConfigMap\r\nmetadata:\r\n  name: " + name + "\r\n ", Kustomize: &KustomizeConfig{Patches: drawPermutation(t, patches)}}})
		}
		NormalizeIntermediateConfig(cfg)
		for i, component := range cfg.Components {
			require.Equal(t, names[i], component.Name)
			require.Equal(t, "kind: ConfigMap\nmetadata:\n  name: "+names[i], component.KubernetesManifest.Manifest)
			require.Equal(t, patches, component.KubernetesManifest.Kustomize.Patches)
		}
		encoded, err := json.Marshal(cfg)
		require.NoError(t, err)
		NormalizeIntermediateConfig(cfg)
		again, err := json.Marshal(cfg)
		require.NoError(t, err)
		require.Equal(t, encoded, again, "normalization must be idempotent")
		var decoded AppConfig
		decoder := json.NewDecoder(bytes.NewReader(encoded))
		decoder.UseNumber()
		require.NoError(t, decoder.Decode(&decoded))
		require.Equal(t, large, decoded.Inputs.Inputs[0].Default, "supported intermediate JSON must preserve exact numbers")
		require.False(t, decoded.Diff(cfg).DirectSummary().HasChanged)
		decoded.Components[0].KubernetesManifest.Manifest += "\ndata:\n  changed: true"
		require.True(t, decoded.Diff(cfg).DirectSummary().HasChanged, "round trips must not hide subsequent content changes")
	})
}

func TestRunbookParse_LegacyRoundTrip(t *testing.T) {
	name := "api"
	for label, dependents := range map[string]bool{"without dependents": false, "with dependents": true} {
		t.Run(label, func(t *testing.T) {
			legacy := &RunbookConfig{Name: "release", Steps: []*RunbookStepConfig{{Name: "deploy", Type: RunbookStepTypeDeployLegacy, ComponentName: name, DeployDependenciesLegacy: dependents}}}
			canonical := &RunbookConfig{Name: "release", Steps: []*RunbookStepConfig{{Name: "deploy", Type: RunbookStepTypeComponentDeploy, ComponentName: name, DeployDependents: dependents}}}
			require.NoError(t, legacy.parse())
			require.NoError(t, canonical.parse())
			require.Equal(t, RunbookStepTypeComponentDeploy, legacy.Steps[0].Type)
			require.Equal(t, dependents, legacy.Steps[0].DeployDependents)
			old, next := &AppConfig{Runbooks: []*RunbookConfig{canonical}}, &AppConfig{Runbooks: []*RunbookConfig{legacy}}
			require.False(t, next.Diff(old).DirectSummary().HasChanged)
			require.NoError(t, legacy.parse())
			require.False(t, next.Diff(old).DirectSummary().HasChanged, "repeated parsing must preserve canonical meaning")
			legacy.Steps[0].PlanOnly = true
			require.True(t, next.Diff(old).DirectSummary().HasChanged)
		})
	}
}
