package configs

type PrivateDockerPullBuild struct {
	Plugin string `hcl:"plugin,label"`

	Image string `hcl:"image"`
	Tag   string `hcl:"tag"`

	EncodedAuth       string `hcl:"encoded_auth"`
	DisableEntrypoint bool   `hcl:"disable_entrypoint,optional"`
}
