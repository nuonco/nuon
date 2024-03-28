package configs

type AWSECRRegistry struct {
	Plugin string `hcl:"plugin,label"`

	Repository string `hcl:"repository"`
	Tag        string `hcl:"tag"`
	Region     string `hcl:"region,optional"`
}

type DockerRegistryAuth struct {
	Hostname      string `hcl:"hostname,optional"`
	Username      string `hcl:"username,optional"`
	Password      string `hcl:"password,optional"`
	Email         string `hcl:"email,optional"`
	Auth          string `hcl:"auth,optional"`
	ServerAddress string `hcl:"serverAddress,optional"`
	IdentityToken string `hcl:"identityToken,optional"`
	RegistryToken string `hcl:"registryToken,optional"`
}

type DockerRegistry struct {
	Plugin string `hcl:"plugin,label"`

	// Image is the name of the image plus tag that the image will be pushed as.
	Image string `hcl:"image,attr"`

	// Tag is the tag to apply to the image.
	Tag string `hcl:"tag,attr"`

	// Local if true will not push this image to a remote registry.
	Local bool `hcl:"local,optional"`

	// Authenticates to private registry
	Auth *DockerRegistryAuth `hcl:"auth,block"`

	// The docker specific encoded authentication string to use to talk to the registry.
	EncodedAuth string `hcl:"encoded_auth,optional"`

	// Insecure indicates if the registry should be accessed via http rather than https
	Insecure bool `hcl:"insecure,optional"`

	// Username is the username to use for authentication on the registry.
	Username string `hcl:"username,optional"`

	// Password is the authentication information assocated with username.
	Password string `hcl:"password,optional"`
}
