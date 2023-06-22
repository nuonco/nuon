package configs

// PublicDockerPullBuild is used to pull a public docker image
type PublicDockerPullBuild struct {
	Plugin string `hcl:"plugin,label"`

	Image             string `hcl:"image"`
	Tag               string `hcl:"tag"`
	DisableEntrypoint bool   `hcl:"disable_entrypoint,optional"`
}
