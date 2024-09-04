package db

import "go.uber.org/fx"

func AsPSQL(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"psql"`, `name:"dbs"`),
	)
}

func AsCH(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"ch"`, `name:"dbs"`),
	)
}

func DBGroupParam(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(`group:"dbs"`),
	)
}
