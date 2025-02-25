package db

import "go.uber.org/fx"

func AsMigrator(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"migrators"`),
	)
}

func DBMigratorParam(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(`group:"migrators"`),
	)
}

func DBGroupParam(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(`group:"dbs"`),
	)
}
