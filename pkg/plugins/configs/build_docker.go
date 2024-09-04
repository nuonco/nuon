package configs

// DockerBuild is used to reference a docker build, usually in an ODR
type DockerBuild struct {
	Plugin string `hcl:"plugin,label"`

	// Controls whether or not the image should be build with buildkit or docker v1
	UseBuildKit bool `hcl:"buildkit,optional"`

	// The name/path to the Dockerfile if it is not the root of the project
	Dockerfile string `hcl:"dockerfile,optional"`

	// Controls the passing of platform flag variables
	Platform string `hcl:"platform,optional"`

	// Controls the passing of build time variables
	BuildArgs map[string]*string `hcl:"build_args,optional"`

	// Controls the passing of build context
	Context string `hcl:"context,optional"`

	// Authenticates to private registry for pulling
	Auth *Auth `hcl:"auth,block"`

	// Controls the passing of the target stage
	Target string `hcl:"target,optional"`

	// Disable the build cache
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
