package configs

// OCIArchiveAuth is used to authenticate with an OCI archive
type OciArchiveAuth struct {
	Username    string `hcl:"username" validate:"required"`
	AuthToken   string `hcl:"auth_token" validate:"required"`
	RegistryURL string `hcl:"registry_url" validate:"required"`
}

// OCISyncBuild is used to sync an oci artifact into a local registry
type OCISyncBuild struct {
	Plugin string `hcl:"plugin,label"`

	Image string         `hcl:"image"`
	Tag   string         `hcl:"tag"`
	Auth  OciArchiveAuth `hcl:"auth,block"`
}
