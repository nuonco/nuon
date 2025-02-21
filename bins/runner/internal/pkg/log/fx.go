package log

import "go.uber.org/fx"

func AsSystemLogger(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"system"`),
	)
}

func AsDevLogger(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"dev"`),
	)
}
