package registry

type AccessInfoAuth struct {
	// base 64 version
	Encoded string

	// required if encoded not set
	Username string
	Password string

	// set if the authentication should use a custom endpoint
	ServerAddress string
}

type AccessInfo struct {
	Image    string
	Insecure bool

	Auth *AccessInfoAuth
}
