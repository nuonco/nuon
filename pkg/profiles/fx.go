package profiles

import "go.uber.org/fx"

func Module(options ...ProfilerOptions) fx.Option {
	opts := LoadOptionsFromEnv()

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

func EnvModule() fx.Option {
	return Module(LoadOptionsFromEnv())
}
