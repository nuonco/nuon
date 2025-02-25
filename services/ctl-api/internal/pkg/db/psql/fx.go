package psql

import "go.uber.org/fx"

func AsPSQL(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"psql"`, `name:"dbs"`),
	)
}
