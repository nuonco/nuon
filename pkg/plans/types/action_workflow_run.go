package plantypes

type ActionWorkflowRunPlan struct {
	ID        string `json:"id"`
	InstallID string `json:"install_id"`

	Attrs map[string]string `json:"attrs"`

	Steps   []*ActionWorkflowRunStepPlan `json:"steps"`
	EnvVars map[string]string            `json:"env_vars"`
}

type ActionWorkflowRunStepPlan struct {
	ID string `json:"run_id"`

	Attrs                      map[string]string `json:"attrs"`
	InterpolatedEnvVars        map[string]string `json:"interpolated_env_vars"`
	GitSource                  *GitSource        `json:"git_source"`
	InterpolatedInlineContents string            `json:"interpolated_inline_contents"`
	InterpolatedCommand        string            `json:"interpolated_command"`
}
