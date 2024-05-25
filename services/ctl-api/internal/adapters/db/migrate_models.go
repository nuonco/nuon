package db

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type joinTable struct {
	model     interface{}
	field     string
	joinTable interface{}
}

func (a *AutoMigrate) migrateModels(ctx context.Context) error {
	a.l.Info("running auto migrate")

	// NOTE: we have to register all join tables manually, since we use soft deletes + custom ID functions
	joinTables := []joinTable{
		{
			&app.Component{},
			"Dependencies",
			&app.ComponentDependency{},
		},
		{
			&app.Installer{},
			"Apps",
			&app.InstallerApp{},
		},
	}
	for _, joinTable := range joinTables {
		if err := a.db.WithContext(ctx).SetupJoinTable(joinTable.model, joinTable.field, joinTable.joinTable); err != nil {
			return fmt.Errorf("unable to create join table: %w", err)
		}
	}

	models := allModels()
	for _, model := range models {
		if err := a.db.WithContext(ctx).AutoMigrate(model); err != nil {
			return err
		}
	}

	return nil
}
