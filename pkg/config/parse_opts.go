package config

type ParseOptions struct {
	RootDir string `json:"RootDir"`
}

type ParseOption func(*ParseOptions)

func WithRootDir(rootDir string) ParseOption {
	return func(opts *ParseOptions) {
		opts.RootDir = rootDir
	}
}

func parseOptions(opts ...ParseOption) ParseOptions {
	var cfg ParseOptions
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
