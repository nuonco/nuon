package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetTableMetricsRequest struct {
	Typ string `validate:"required"`
}

// @temporal-gen activity
// @by-id Typ
func (a *Activities) GetTableMetrics(ctx context.Context, req GetTableMetricsRequest) ([]app.TableSize, error) {
	var db *gorm.DB
	switch req.Typ {
	case "psql":
		db = a.db
	case "ch":
		db = a.chDB
	}

	return a.getTableSizes(ctx, db)
}

func (a *Activities) getTableSizes(ctx context.Context, db *gorm.DB) ([]app.TableSize, error) {
	var tables []app.TableSize

	res := db.WithContext(ctx).
		Find(&tables)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get table sizes: %w", res.Error)
	}

	return tables, nil
}
