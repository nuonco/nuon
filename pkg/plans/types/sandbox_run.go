package plantypes

import (
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"
	"github.com/powertoolsdev/mono/pkg/types/state"
)

type TerraformRunType string

const (
	TerraformRunTypeApply   TerraformRunType = "apply"
	TerraformRunTypeDestroy TerraformRunType = "destroy"

	// TODO(jm): support when we use plans
	// TerraformRunTypeApplyWithPlan   TerraformRunType = "apply-with-plan"
	// TerraformRunTypeDestroyWithPlan TerraformRunType = "destroy-with-plan"
	// TerraformRunTypeCreateApplyPlan TerraformRunType = "destroy-with-plan"
)

type TerraformBackend struct {
	WorkspaceID string `validate:"required"`
}

type TerraformLocalArchive struct {
	Path string `json:"local_archive"`
}

type SandboxRunPlan struct {
	InstallID   string `json:"install_id"`
	AppID       string `json:"app_id"`
	AppConfigID string `json:"app_config_id"`

	Vars             map[string]any           `json:"vars"`
	EnvVars          map[string]string        `json:"env_vars"`
	VarsFiles        []string                 `json:"vars_files"`
	GitSource        *GitSource               `json:"git_source"`
	LocalArchive     *TerraformLocalArchive   `json:"local_archive"`
	TerraformBackend *TerraformBackend        `json:"terraform_backend"`
	AzureAuth        *azurecredentials.Config `json:"azure_auth"`
	AWSAuth          *awscredentials.Config   `json:"aws_auth"`
	Hooks            *TerraformDeployHooks    `json:"hooks"`

	Policies map[string]string `json:"policies"`

	State *state.State `json:"state"`
}

type TerraformDeployHooks struct {
	Enabled bool               `hcl:"enabled"`
	EnvVars map[string]string  `hcl:"env_vars"`
	RunAuth credentials.Config `hcl:"run_auth,block"`
}
