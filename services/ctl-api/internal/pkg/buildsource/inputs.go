package buildsource

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"github.com/nuonco/nuon/pkg/config"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

func ComponentInputs(c *app.ComponentConfigConnection) (*plantypes.BuildPlan, error) {
	if c == nil {
		return nil, fmt.Errorf("component config connection is nil")
	}

	plan := &plantypes.BuildPlan{}
	switch {
	case c.TerraformModuleComponentConfig != nil:
		plan.TerraformBuildPlan = &plantypes.TerraformBuildPlan{
			VendorProviders:  true,
			TerraformVersion: sourceidentity.TerraformVersion(c.TerraformModuleComponentConfig.Version),
		}
	case c.OpentofuModuleComponentConfig != nil:
		plan.OpentofuBuildPlan = &plantypes.OpentofuBuildPlan{
			VendorProviders: true,
			OpenTofuVersion: opentofuVersion(c.OpentofuModuleComponentConfig.Version),
		}
	case c.HelmComponentConfig != nil:
		cfg := c.HelmComponentConfig.HelmConfig
		if cfg == nil {
			return nil, fmt.Errorf("helm config is nil")
		}
		helmPlan := &plantypes.HelmBuildPlan{
			ValuesFiles: slices.Clone(cfg.ValuesFiles),
			Values:      make([]plantypes.HelmValue, 0, len(cfg.Values)),
		}
		if repo := cfg.HelmRepoConfig; repo != nil {
			helmPlan.HelmRepoConfig = &config.HelmRepoConfig{
				Chart:   repo.Chart,
				RepoURL: repo.RepoURL,
				Version: repo.Version,
			}
		}
		for name, value := range cfg.Values {
			if value != nil {
				helmPlan.Values = append(helmPlan.Values, plantypes.HelmValue{
					Name: name, Value: *value, Type: "auto",
				})
			}
		}
		sort.Slice(helmPlan.Values, func(i, j int) bool {
			return helmPlan.Values[i].Name < helmPlan.Values[j].Name
		})
		plan.HelmBuildPlan = helmPlan
	case c.KubernetesManifestComponentConfig != nil:
		cfg := c.KubernetesManifestComponentConfig
		manifestPlan := &plantypes.KubernetesManifestBuildPlan{SourceType: "inline", InlineManifest: cfg.Manifest}
		if cfg.Kustomize != nil && cfg.Kustomize.Path != "" {
			manifestPlan = &plantypes.KubernetesManifestBuildPlan{
				SourceType:    "kustomize",
				KustomizePath: cfg.Kustomize.Path,
				KustomizeConfig: &plantypes.KustomizeBuildConfig{
					Patches:        slices.Clone(cfg.Kustomize.Patches),
					EnableHelm:     cfg.Kustomize.EnableHelm,
					LoadRestrictor: cfg.Kustomize.LoadRestrictor,
				},
			}
		}
		plan.KubernetesManifestBuildPlan = manifestPlan
	case c.PulumiComponentConfig != nil:
		plan.PulumiBuildPlan = &plantypes.PulumiBuildPlan{}
	case c.ExternalImageComponentConfig != nil:
		cfg := c.ExternalImageComponentConfig
		plan.ContainerImagePullPlan = &plantypes.ContainerImagePullPlan{
			Image:        cfg.ImageURL,
			Tag:          cfg.Tag,
			UpdatePolicy: cfg.UpdatePolicy,
			Verification: cfg.Verification,
		}
	default:
		return nil, fmt.Errorf("unsupported or unloaded component build config %q", c.Type)
	}
	return plan, nil
}

func opentofuVersion(version string) string {
	if version == "" {
		return build.DefaultOpentofuVersion
	}
	return version
}

func ComponentChecksum(c *app.ComponentConfigConnection) (string, error) {
	plan, err := ComponentInputs(c)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(plan)
	if err != nil {
		return "", fmt.Errorf("unable to encode component build inputs: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func SandboxInputs(c *app.AppSandboxConfig) *plantypes.BuildPlan {
	plan := &plantypes.BuildPlan{}
	switch c.Type {
	case config.AppSandboxTypeOpentofu:
		plan.OpentofuBuildPlan = &plantypes.OpentofuBuildPlan{
			VendorProviders: true,
			OpenTofuVersion: opentofuVersion(c.OpenTofuVersion),
		}
	default:
		plan.TerraformBuildPlan = &plantypes.TerraformBuildPlan{
			VendorProviders:  true,
			TerraformVersion: sourceidentity.TerraformVersion(c.TerraformVersion),
		}
	}
	return plan
}

func SandboxPlan(c *app.AppSandboxConfig, labels map[string]string) *plantypes.BuildPlan {
	plan := SandboxInputs(c)
	if plan.TerraformBuildPlan != nil {
		plan.TerraformBuildPlan.Labels = labels
	} else if plan.OpentofuBuildPlan != nil {
		plan.OpentofuBuildPlan.Labels = labels
	}
	return plan
}

func SandboxChecksum(c *app.AppSandboxConfig) string {
	return sourceidentity.Hash(SandboxInputs(c))
}
