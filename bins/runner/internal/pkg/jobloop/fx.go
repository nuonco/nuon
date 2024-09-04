package jobloop

import (
	"go.uber.org/fx"
)

func AsJobLoop(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"job_loops"`),
		fx.As(new(JobLoop)),
	)
}

func WithJobLoops(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(`group:"job_loops"`),
	)
}
