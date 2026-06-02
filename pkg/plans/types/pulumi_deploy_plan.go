package plantypes

import (
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/pkg/types/state"
)

type PulumiDeployPlan struct {
	Config  map[string]string `json:"config" faker:"-"`
	EnvVars map[string]string `json:"env_vars"`

	Runtime       string `json:"runtime"`
	PulumiVersion string `json:"pulumi_version"`
	StackName     string `json:"stack_name"`

	// Reuse workspace concept for state storage
	WorkspaceID string `json:"workspace_id"`

	AzureAuth *azurecredentials.Config `json:"azure_auth"`
	AWSAuth   *awscredentials.Config   `json:"aws_auth"`
	GCPAuth   *gcpcredentials.Config   `json:"gcp_auth"`

	ClusterInfo *kube.ClusterInfo `json:"cluster_info,block"`

	State *state.State `json:"state" faker:"-"`

	// Destroy indicates this is a teardown operation (pulumi destroy instead of up)
	Destroy bool `json:"destroy"`

	UpdatePlans bool `json:"update_plans,omitempty"`

	PlanJSON []byte `json:"plan_json"`
}
