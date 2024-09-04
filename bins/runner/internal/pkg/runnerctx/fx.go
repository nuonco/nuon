package runnerctx

import "go.uber.org/fx"

func AsCancelFn(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"cancel_fn"`),
	)
}
