package configs

type DockerRefBuild struct {
	Plugin string `hcl:"plugin,label"`

	Image string `hcl:"image"`
	Tag   string `hcl:"tag"`
}
