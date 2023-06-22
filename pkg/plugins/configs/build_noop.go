package configs

func DefaultNoopBuild() Build[NoopBuild, Registry[NoopRegistry]] {
	return Build[NoopBuild, Registry[NoopRegistry]]{
		Use: NoopBuild{
			Plugin: "noop",
		},
		Registry: Registry[NoopRegistry]{
			Use: NoopRegistry{
				Plugin: "noop",
			},
		},
	}
}

type NoopBuild struct {
	Plugin string `hcl:"plugin,label"`
}
