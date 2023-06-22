package configs

type DockerRegistry struct {
	Plugin string `hcl:"plugin,label"`

	Repository string `hcl:"repository"`
	Tag        string `hcl:"tag"`
	Region     string `hcl:"region,optional"`
}
