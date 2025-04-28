package plantypes

import (
	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"
	"github.com/powertoolsdev/mono/pkg/types/state"
)

type TerraformDeployPlan struct {
	Vars      map[string]any    `json:"vars" faker:"-"`
	EnvVars   map[string]string `json:"env_vars"`
	VarsFiles []string          `json:"vars_files"`

	TerraformBackend *TerraformBackend        `json:"terraform_backend"`
	AzureAuth        *azurecredentials.Config `json:"azure_auth"`
	AWSAuth          *awscredentials.Config   `json:"aws_auth"`
	Hooks            *TerraformDeployHooks    `json:"hooks"`

	Policies map[string]string `json:"policies"`

	State *state.State `json:"state" faker:"-"`
}
