package configs

// DockerBuild is used to reference a docker build, usually in an ODR
type DockerBuild struct {
	Plugin string `hcl:"plugin,label"`

	Dockerfile        string `hcl:"dockerfile"`
	DisableEntrypoint bool   `hcl:"disable_entrypoint"`
}
