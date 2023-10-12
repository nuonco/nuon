package configs

type JobDeploy struct {
	Plugin        string            `hcl:"plugin,label"`
	StaticEnvVars map[string]string `hcl:"static_environment,optional"`
	ImageURL      string            `hcl:"image_url"`
	Tag           string            `hcl:"tag"`
	Cmd           string            `hcl:"cmd"`
}
