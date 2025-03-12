package jobloop

import (
	"go.uber.org/fx"
)

// we break up job loops into two groups, job_loops and operations, because we want to
// monitor job_loops but not operations. this means we have to invoke the operations
// job loops on their own though.

func AsJobLoop(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"job_loops"`),
		fx.As(new(JobLoop)),
	)
}

func AsOperationsJobLoop(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"operations"`), // rename later
		fx.As(new(JobLoop)),
	)
}

func WithJobLoops(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(`group:"job_loops"`),
	)
}

func WithOperationsJobLoops(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(`group:"operations"`),
	)
}
