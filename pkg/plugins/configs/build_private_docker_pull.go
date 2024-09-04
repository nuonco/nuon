package configs

// PrivateDockerPullBuild is used to reference a private docker build
type PrivateDockerPullBuild struct {
	Plugin string `hcl:"plugin,label"`

	Image string `hcl:"image"`
	Tag   string `hcl:"tag"`

	// the encoded auth that we use throughout our code looks something like the following,
	EncodedAuth       string `hcl:"encoded_auth"`
	DisableEntrypoint bool   `hcl:"disable_entrypoint,optional"`
}
