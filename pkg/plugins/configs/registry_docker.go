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

	Image string `hcl:"image,attr"`

	Tag string `hcl:"tag,attr"`

	Local bool `hcl:"local,optional"`

	Auth *DockerRegistryAuth `hcl:"auth,block"`

	EncodedAuth string `hcl:"encoded_auth,optional"`

	Insecure bool `hcl:"insecure,optional"`

	Username string `hcl:"username,optional"`

	Password string `hcl:"password,optional"`
}
