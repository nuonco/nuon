package configs

type JobDeploy struct {
	PlanOnly bool              `hcl:"plan_only,optional"`
	Plugin   string            `hcl:"plugin,label"`
	EnvVars  map[string]string `hcl:"env_vars,optional"`
	ImageURL string            `hcl:"image_url"`
	Tag      string            `hcl:"tag"`
	Cmd      []string          `hcl:"cmd,optional"`
	Args     []string          `hcl:"args"`

	JobName        string `hcl:"job_name"`
	ContainerName  string `hcl:"container_name"`
	Namespace      string `hcl:"namespace"`
	ServiceAccount string `hcl:"service_account"`
	RestartPolicy  string `hcl:"restart_policy"`
}
