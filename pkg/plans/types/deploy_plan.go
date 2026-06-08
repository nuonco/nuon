package plantypes

import (
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

type DeployPlan struct {
	InstallID     string `json:"install_id"`
	AppID         string `json:"app_id"`
	AppConfigID   string `json:"app_config_id"`
	ComponentID   string `json:"component_id"`
	ComponentName string `json:"component_name"`

	Src    *configs.OCIRegistryRepository `json:"src_registry" validate:"required"`
	SrcTag string                         `json:"src_tag" validate:"required"`
	// SrcDigest is the manifest digest of the source artifact in the install
	// registry, e.g. "sha256:abc...". Populated for image-type component
	// builds with source identity recorded; empty for
	// non-image builds and legacy image builds. When non-empty, runners
	// should prefer this over SrcTag for content-addressed pulls and for
	// rendering digest-pinned image references in pod specs.
	SrcDigest string `json:"src_digest,omitzero,omitempty"`

	HelmDeployPlan               *HelmDeployPlan               `json:"helm"`
	TerraformDeployPlan          *TerraformDeployPlan          `json:"terraform"`
	KubernetesManifestDeployPlan *KubernetesManifestDeployPlan `json:"kubernetes_manifest"`
	NoopDeployPlan               *NoopDeployPlan               `json:"noop"`
	PulumiDeployPlan             *PulumiDeployPlan             `json:"pulumi"`

	// The following field is for applying a plan that is already save
	ApplyPlanContents string `json:"apply_plan_contents"`
	// This field is for storing a human legible plan or corollary representation
	ApplyPlanDisplay string `json:"apply_plan_display"`

	SandboxMode *SandboxMode `json:"sandbox_mode,omitzero,omitempty"`
}
