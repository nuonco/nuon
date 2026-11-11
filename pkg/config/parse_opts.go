package config

import "github.com/nuonco/nuon/pkg/sourcefs"

type ParseOptions struct {
	RootDir string
	Source  *sourcefs.FS
}

type ParseOption func(*ParseOptions)

func WithRootDir(rootDir string) ParseOption {
	return func(opts *ParseOptions) {
		opts.RootDir = rootDir
	}
}

// WithSource borrows a repository filesystem; the caller retains ownership.
func WithSource(source *sourcefs.FS) ParseOption {
	return func(opts *ParseOptions) {
		opts.Source = source
	}
}

func parseOptions(opts ...ParseOption) ParseOptions {
	var cfg ParseOptions
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
