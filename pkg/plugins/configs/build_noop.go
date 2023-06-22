package configs

func DefaultNoopBuild() Build[NoopBuild, Registry[DockerRegistry]] {
	return Build[NoopBuild, Registry[DockerRegistry]]{
		Use: NoopBuild{
			Plugin: "docker-pull",

			Image:             "hashicorpdemoapp/public-api",
			Tag:               "v0.0.5",
			DisableEntrypoint: true,
		},
		Registry: Registry[DockerRegistry]{
			Use: DockerRegistry{
				Plugin:     "aws-ecr",
				Repository: "nuon.local",
				Tag:        "latest",
			},
		},
	}
}

type NoopBuild struct {
	Plugin string `hcl:"plugin,label"`

	Image             string `hcl:"image"`
	Tag               string `hcl:"tag"`
	DisableEntrypoint bool   `hcl:"disable_entrypoint"`
}
