package migrations

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
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

func (m *Migrator) dropViews(ctx context.Context) error {
	names := []string{}
	for _, model := range m.models {
		if vm, ok := m.toViewModel(model); ok {
			for _, view := range vm.Views(m.db) {
				if view.AlwaysReapply {
					names = append(names, m.db.Statement.Quote(view.Name))
				}
			}
		}
	}
	if len(names) == 0 {
		return nil
	}
	// Views registered later may depend on earlier ones, so drop dependents first.
	slices.Reverse(names)
	err := m.db.WithContext(ctx).Exec("DROP VIEW IF EXISTS " + strings.Join(names, ", ")).Error
	if err == nil {
		return nil
	}
	// PostgreSQL names the refused view in the message but keeps the object that
	// blocked it in DETAIL; surface the detail so the operator knows what to remove.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Detail != "" {
		return errors.Wrap(err, pgErr.Detail)
	}
	return err
}

// applyViewsTransactionally runs the drop and recreate inside one transaction
// with a bounded lock wait, so a slow view reader cannot stall startup for the
// full migration timeout. Deadlocks and lock waits are retried with backoff,
// mirroring the lock_timeout pattern in migration 144.
func (m *Migrator) applyViewsTransactionally(ctx context.Context, apply func() error) error {
	for attempt := 0; ; attempt++ {
		err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("SET LOCAL lock_timeout = '2s'").Error; err != nil {
				return err
			}
			orig := m.db
			m.db = tx
			defer func() { m.db = orig }()
			if err := m.dropViews(ctx); err != nil {
				return errors.Wrap(err, "unable to drop registered views")
			}
			return apply()
		})
		var pgErr *pgconn.PgError
		if attempt == 2 || !errors.As(err, &pgErr) || (pgErr.Code != "55P03" && pgErr.Code != "40P01") {
			return err
		}
		m.l.Warn("retrying views migration after lock contention",
			zap.Int("attempt", attempt+1),
			zap.Error(err))
		timer := time.NewTimer(100 * time.Millisecond * time.Duration(1<<attempt))
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
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
	if view.AlwaysReapply && m.db.Dialector.Name() != "postgres" {
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
