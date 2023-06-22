package configs

func DefaultNoopBuild() Build[NoopBuild, Registry[NoopRegistry]] {
	return Build[NoopBuild, Registry[NoopRegistry]]{
		Use: NoopBuild{
			Plugin: "noop",
		},
		Registry: Registry[NoopRegistry]{
			Use: NoopRegistry{
				Plugin: "aws-ecr",
			},
		},
	}
}

type NoopBuild struct {
	Plugin string `hcl:"plugin,label"`
}
