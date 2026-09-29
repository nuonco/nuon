package configs

type OciArchiveAuth struct {
	Username    string `hcl:"username" validate:"required"`
	AuthToken   string `hcl:"auth_token" validate:"required"`
	RegistryURL string `hcl:"registry_url" validate:"required"`
}

type OCISyncBuild struct {
	Plugin string `hcl:"plugin,label"`

	Image  string                `hcl:"image"`
	Tag    string                `hcl:"tag"`
	Source OCIRegistryRepository `hcl:"source,block"`
}
