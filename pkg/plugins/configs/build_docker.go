package configs

type DockerBuild struct {
	Plugin string `hcl:"plugin,label"`

	UseBuildKit bool `hcl:"buildkit,optional"`

	Dockerfile string `hcl:"dockerfile,optional"`

	Platform string `hcl:"platform,optional"`

	BuildArgs map[string]*string `hcl:"build_args,optional"`

	Context string `hcl:"context,optional"`

	Auth *Auth `hcl:"auth,block"`

	Target string `hcl:"target,optional"`

	NoCache bool `hcl:"no_cache,optional"`
}

type Auth struct {
	Hostname      string `hcl:"hostname,optional"`
	Username      string `hcl:"username,optional"`
	Password      string `hcl:"password,optional"`
	Email         string `hcl:"email,optional"`
	Auth          string `hcl:"auth,optional"`
	EncodedAuth   string `hcl:"auth,optional"`
	ServerAddress string `hcl:"serverAddress,optional"`
	IdentityToken string `hcl:"identityToken,optional"`
	RegistryToken string `hcl:"registryToken,optional"`
}
