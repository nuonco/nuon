package slog

import (
	"go.uber.org/fx"
)

func AsSystemProvider(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"system"`),
	)
}
