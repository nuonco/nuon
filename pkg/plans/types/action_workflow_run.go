package plantypes

import (
	"time"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/pkg/kube"
)

type ActionWorkflowRunPlan struct {
	ID        string `json:"id"`
	InstallID string `json:"install_id"`

	Attrs map[string]string `json:"attrs"`

	Steps           []*ActionWorkflowRunStepPlan `json:"steps"`
	BuiltinEnvVars  map[string]string            `json:"builtin_env_vars"`
	OverrideEnvVars map[string]string            `json:"override_env_vars"`
	Timeout         time.Duration                `json:"timeout,omitempty" swaggertype:"primitive,integer"`

	// optional fields based on the configuration
	ClusterInfo *kube.ClusterInfo        `json:"cluster_info,block"`
	AWSAuth     *awscredentials.Config   `json:"aws_auth,omitempty"`
	AzureAuth   *azurecredentials.Config `json:"azure_auth,omitempty"`
	GCPAuth     *gcpcredentials.Config   `json:"gcp_auth,omitempty"`

	MinSandboxMode
}

type ActionWorkflowRunStepPlan struct {
	ID string `json:"run_id"`

	Attrs                      map[string]string `json:"attrs"`
	InterpolatedEnvVars        map[string]string `json:"interpolated_env_vars"`
	GitSource                  *GitSource        `json:"git_source"`
	InterpolatedInlineContents string            `json:"interpolated_inline_contents"`
	InterpolatedCommand        string            `json:"interpolated_command"`
}
