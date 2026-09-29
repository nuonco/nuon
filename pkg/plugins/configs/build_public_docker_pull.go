package configs

type PublicDockerPullBuild struct {
	Plugin string `hcl:"plugin,label"`

	Image             string `hcl:"image"`
	Tag               string `hcl:"tag"`
	DisableEntrypoint bool   `hcl:"disable_entrypoint,optional"`
}
