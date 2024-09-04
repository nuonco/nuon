package worker

import "go.uber.org/fx"

func AsWorker(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"worker"`),
		fx.As(new(Worker)),
	)
}

func WithWorkers(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(`group:"worker"`),
	)
}
