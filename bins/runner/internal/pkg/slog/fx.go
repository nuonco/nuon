package slog

import (
	"go.uber.org/fx"
)

func AsSystemLogger(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"system"`),
	)
}

func AsSystemProvider(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"system"`),
	)
}

func AsJobProvider(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"job"`),
	)
}
