package migrations

import "context"

type Migration struct {
	Name     string
	Fn       func(context.Context) error
	Disabled bool
}

func (a *Migrations) GetAll() []Migration {
	return []Migration{
		{
			Name:     "001-sql-example",
			Fn:       a.migration001ExampleSQL,
			Disabled: true,
		},
		{
			Name:     "002-model-migration",
			Fn:       a.migration002ExampleModel,
			Disabled: true,
		},
		{
			Name:     "003-seed",
			Fn:       a.migration003Seed,
			Disabled: false,
		},
		{
			Name: "041-app-config-view",
			Fn:   a.migration041AppConfigVersions,
		},
		{
			Name: "043-component-config-connections-view",
			Fn:   a.migration043ComponentConfigConnectionsView,
		},
		{
			Name: "044-installs-view",
			Fn:   a.migration044InstallsView,
		},
	}
}
