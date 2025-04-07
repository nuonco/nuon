package profiles

import "go.uber.org/fx"

// Module provides the profiler as an fx module
func Module(options ...ProfilerOptions) fx.Option {
	opts := LoadOptionsFromEnv()

	// If options are explicitly provided, they take precedence over env vars
	if len(options) > 0 {
		opts = options[0]
	}

	return fx.Module(
		"profiler",
		fx.Provide(func() ProfilerOptions {
			return opts
		}),
		fx.Invoke(RegisterProfiler),
	)
}

// EnvModule provides the profiler as an fx module using only environment config
func EnvModule() fx.Option {
	return Module(LoadOptionsFromEnv())
}
