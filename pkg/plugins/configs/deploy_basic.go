package configs

type Registry struct {
	Name string `hcl:"name,label"`

	Repository string `hcl:"repository,label"`
	Tag        string `hcl:"tag,label"`
	Region     string `hcl:"region,label"`
}

// DockerBuild is used to reference a docker build, usually in an ODR
type DockerBuild struct {
	Name string `hcl:"name,label"`

	Dockerfile string   `hcl:"dockerfile"`
	Registry   Registry `hcl:"registry,block"`
}
