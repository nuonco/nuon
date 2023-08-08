package configs

func DefaultNoopDeploy() Deploy[NoopDeploy] {
	return Deploy[NoopDeploy]{
		Use: NoopDeploy{
			Plugin: "noop",
		},
	}
}

type NoopDeploy struct {
	Plugin string `hcl:"plugin,label"`
}
