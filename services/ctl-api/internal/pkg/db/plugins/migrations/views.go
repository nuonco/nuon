package migrations

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

type View struct {
	Name string

	SQL string
	DB  *gorm.DB

	// Controls whether a view is always reapplied (by dropping it and recreating it)
	AlwaysReapply bool
}

type viewModel interface {
	Views(*gorm.DB) []View
}

func (m *Migrator) toViewModel(obj any) (viewModel, bool) {
	jtm, ok := obj.(viewModel)
	return jtm, ok
}

func (m *Migrator) applyViews(ctx context.Context, obj any) error {
	im, ok := m.toViewModel(obj)
	if !ok {
		return nil
	}

	for _, idx := range im.Views(m.db) {
		if err := m.applyView(ctx, obj, idx); err != nil {
			return MigrationErr{
				Model: plugins.TableName(m.db, obj),
				Name:  "views",
				Err:   err,
			}
		}
	}

	return nil
}

func (m *Migrator) applyView(ctx context.Context, obj any, view View) error {
	if view.AlwaysReapply {
		dropSQLTmpl := m.opts.DropViewSQLTmpl
		dropSQL := fmt.Sprintf(dropSQLTmpl, view.Name)

		m.l.Debug("dropping view",
			zap.String("name", plugins.TableName(m.db, obj)),
			zap.String("sql", dropSQL),
		)
		if res := m.db.WithContext(ctx).
			Exec(dropSQL); res.Error != nil {
			return errors.Wrap(res.Error, "unable to drop view "+view.Name)
		}
	}

	applySQLTmpl := m.opts.CreateViewSQLTmpl
	applySQL := fmt.Sprintf(applySQLTmpl, view.Name, view.SQL)
	m.l.Debug("creating view",
		zap.String("name", plugins.TableName(m.db, obj)),
		zap.String("sql", applySQL),
	)
	if res := m.db.WithContext(ctx).
		Exec(applySQL); res.Error != nil {
		return errors.Wrap(res.Error, "unable to create view "+view.Name)
	}

	return nil
}
