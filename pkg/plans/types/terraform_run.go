package plantypes

import (
	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"
	"github.com/powertoolsdev/mono/pkg/types/state"
)

type TerraformRunPlan struct {
	InstallID   string `json:"install_id"`
	AppID       string `json:"app_id"`
	AppConfigID string `json:"app_config_id"`

	Vars             map[string]any           `json:"vars"`
	EnvVars          map[string]string        `json:"env_vars"`
	GitSource        *GitSource               `json:"git_source"`
	LocalArchive     *TerraformLocalArchive   `json:"local_archive"`
	TerraformBackend *TerraformBackend        `json:"terraform_backend"`
	AzureAuth        *azurecredentials.Config `json:"azure_auth"`
	AWSAuth          *awscredentials.Config   `json:"aws_auth"`
	Hooks            *TerraformDeployHooks    `json:"hooks"`

	Policies map[string]string `json:"policies"`

	State *state.State `json:"state"`
}
