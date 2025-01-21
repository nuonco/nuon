package api

import (
	"go.uber.org/fx"
)


func AsAPI(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"api"`),
	)
}

func APIGroupParam(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(`group:"api"`),
	)
}
