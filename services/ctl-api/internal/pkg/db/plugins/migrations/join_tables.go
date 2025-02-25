package migrations

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

type JoinTable struct {
	Field     string
	JoinTable interface{}
}

type joinsTableModel interface {
	JoinTables() []JoinTable
}

func (m *Migrator) toJoinTables(obj any) (joinsTableModel, bool) {
	jtm, ok := obj.(joinsTableModel)
	return jtm, ok
}

func (m *Migrator) applyJoinTables(ctx context.Context, obj any) error {
	jtm, ok := m.toJoinTables(obj)
	if !ok {
		return nil
	}

	for _, jt := range jtm.JoinTables() {
		if err := m.applyJoinTable(ctx, obj, jt); err != nil {
			return MigrationErr{
				Model: plugins.TableName(m.db, obj),
				Name:  "join_tables",
				Err:   err,
			}
		}
	}

	return nil
}

func (m *Migrator) applyJoinTable(ctx context.Context, obj any, jt JoinTable) error {
	if err := m.db.
		WithContext(ctx).
		SetupJoinTable(obj, jt.Field, jt.JoinTable); err != nil {
		return errors.Wrap(err, "unable to create join table for "+jt.Field)
	}

	return nil
}
