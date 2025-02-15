package interceptors

import (
	"go.temporal.io/sdk/interceptor"
	"go.uber.org/fx"
)

func AsInterceptor(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(interceptor.WorkerInterceptor)),
		fx.ResultTags(`group:"interceptors"`),
	)
}
