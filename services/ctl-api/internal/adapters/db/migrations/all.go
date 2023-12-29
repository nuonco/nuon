package migrations

import "context"

type Migration struct {
	Name string
	Fn   func(context.Context) error
}

func (a *Migrations) GetAll() []Migration {
	return []Migration{
		{
			Name: "001-sql-example",
			Fn:   a.migration001ExampleSQL,
		},
		{
			Name: "002-model-migration",
			Fn:   a.migration002ExampleModel,
		},
		{
			Name: "003-seed",
			Fn:   a.migration003Seed,
		},
		{
			Name: "004-fix-install-cascade-constraints",
			Fn:   a.migration004InstallsCascadeInputs,
		},
		{
			Name: "005-component-dependencies-primary-key",
			Fn:   a.migration005ComponentDependencyPrimaryKey,
		},
		{
			Name: "006-component-dependencies-deleted-at-constraint",
			Fn:   a.migration006ComponentDependencyDeletedAtConstraint,
		},
	}
}
