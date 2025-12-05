package plantypes

import (
	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/kube"
)

type ActionWorkflowRunPlan struct {
	ID        string `json:"id"`
	InstallID string `json:"install_id"`

	Attrs map[string]string `json:"attrs"`

	Steps           []*ActionWorkflowRunStepPlan `json:"steps"`
	BuiltinEnvVars  map[string]string            `json:"builtin_env_vars"`
	OverrideEnvVars map[string]string            `json:"override_env_vars"`

	// optional fields based on the configuration
	ClusterInfo *kube.ClusterInfo      `json:"cluster_info,block"`
	AWSAuth     *awscredentials.Config `json:"aws_auth"`

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
