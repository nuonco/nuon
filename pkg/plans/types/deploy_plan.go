package plantypes

import (
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

type DeployPlan struct {
	InstallID     string `json:"install_id"`
	AppID         string `json:"app_id"`
	AppConfigID   string `json:"app_config_id"`
	ComponentID   string `json:"component_id"`
	ComponentName string `json:"component_name"`

	Src    *configs.OCIRegistryRepository `json:"src_registry" validate:"required"`
	SrcTag string                         `json:"src_tag" validate:"required"`

	HelmDeployPlan               *HelmDeployPlan               `json:"helm"`
	TerraformDeployPlan          *TerraformDeployPlan          `json:"terraform"`
	KubernetesManifestDeployPlan *KubernetesManifestDeployPlan `json:"kubernetes_manifest"`
	NoopDeployPlan               *NoopDeployPlan               `json:"noop"`

	// The following field is for applying a plan that is already save
	ApplyPlanContents string `json:"apply_plan_contents"`
	// This field is for storing a human legible plan or corollary representation
	ApplyPlanDisplay string `json:"apply_plan_display"`

	SandboxMode *SandboxMode `json:"sandbox_mode,omitzero,omitempty"`
}
