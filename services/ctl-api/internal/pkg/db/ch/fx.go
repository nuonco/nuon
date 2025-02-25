package ch

import "go.uber.org/fx"

func AsCH(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"ch"`, `name:"dbs"`),
	)
}
