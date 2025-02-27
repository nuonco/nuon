package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

type Index struct {
	Name    string
	Columns []string

	UniqueValue  sql.NullBool
	PrimaryValue sql.NullBool

	Option  string
	Type    string
	Comment string
}

type indexModel interface {
	Indexes(*gorm.DB) []Index
}

func (m *Migrator) toIndexMode(obj any) (indexModel, bool) {
	jtm, ok := obj.(indexModel)
	return jtm, ok
}

func (m *Migrator) getIndexes(ctx context.Context, obj any) ([]gorm.Index, error) {
	indexes, err := m.db.WithContext(ctx).Migrator().GetIndexes(obj)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indexes")
	}

	return indexes, nil
}

func (m *Migrator) toGormIndexes(obj any, idxs []Index) []gorm.Index {
	newIdxs := make([]gorm.Index, 0)
	for _, i := range idxs {
		newIdxs = append(newIdxs, gormIndex{
			idx:   i,
			table: plugins.TableName(m.db, obj),
		})
	}

	return newIdxs
}

func (m *Migrator) applyIndexes(ctx context.Context, obj any) error {
	im, ok := m.toIndexMode(obj)

	if !ok {
		return nil
	}

	expectedIndexes := im.Indexes(m.db)
	expected := toIndexMap(m.toGormIndexes(obj, expectedIndexes))

	existingIndexes, err := m.getIndexes(ctx, obj)
	existing := toIndexMap(existingIndexes)
	if err != nil {
		return MigrationErr{
			Model: plugins.TableName(m.db, obj),
			Name:  "indexes",
			Err:   errors.Wrap(err, "unable to get indexes"),
		}
	}

	toAdd, toDel := generics.DiffMaps(expected, existing)

	for _, idx := range toAdd {
		if err := m.applyIndex(ctx, obj, idx.(gormIndex).idx); err != nil {
			return MigrationErr{
				Model: plugins.TableName(m.db, obj),
				Name:  "indexes",
				Err:   err,
			}
		}
	}

	for _, idx := range toDel {
		if err := m.deleteIndex(ctx, obj, idx); err != nil {
			return MigrationErr{
				Model: plugins.TableName(m.db, obj),
				Name:  "indexes",
				Err:   err,
			}
		}
	}

	return nil
}

// BuildIndexOptionsInterface build index options interface
type BuildIndexOptionsInterface interface {
	BuildIndexOptions([]schema.IndexOption, *gorm.Statement) []interface{}
}

func (m *Migrator) applyIndex(ctx context.Context, obj any, idx Index) error {
	if m.db.WithContext(ctx).Migrator().HasIndex(obj, idx.Name) {
		m.l.Debug("index already exists",
			zap.String("model", plugins.TableName(m.db, obj)),
			zap.String("index", idx.Name))
		return nil
	}

	args := []any{
		clause.Column{Name: idx.Name},
		clause.Table{Name: plugins.TableName(m.db, obj)},
	}

	columns := ""
	for i, col := range idx.Columns {
		if i > 0 {
			columns += ", "
		}
		columns += col
	}
	args = append(args, clause.Expr{SQL: columns})

	tmpl := m.opts.CreateIndexTmpl
	if idx.UniqueValue.Valid && idx.UniqueValue.Bool {
		tmpl = m.opts.CreateUniqueIndexTmpl
	}
	if idx.PrimaryValue.Valid && idx.PrimaryValue.Bool {
		tmpl = m.opts.CreatePKIndexTmpl
	}

	if idx.Type != "" {
		tmpl += " USING " + idx.Type
	}

	if idx.Comment != "" {
		tmpl += fmt.Sprintf(" COMMENT '%s'", idx.Comment)
	}

	if idx.Option != "" {
		tmpl += " " + idx.Option
	}

	return m.db.WithContext(ctx).Exec(tmpl, args...).Error
}

func (m *Migrator) deleteIndex(ctx context.Context, obj any, idx gorm.Index) error {
	if !m.allowDestroy {
		m.l.Warn("skipping deleting index",
			zap.String("name", idx.Name()),
			zap.String("model", plugins.TableName(m.db, obj)),
		)

		return nil
	}

	m.l.Error("deleting index",
		zap.String("name", idx.Name()),
		zap.String("model", plugins.TableName(m.db, obj)),
	)

	args := []any{
		clause.Column{Name: idx.Name()},
	}
	return m.db.Exec(m.opts.DropIndexTmpl, args...).Error
}
