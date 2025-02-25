package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetPSQLTableMetricsRequest struct{}

// @temporal-gen activity
func (a *Activities) GetPSQLTableMetrics(ctx context.Context, req GetPSQLTableMetricsRequest) ([]app.PSQLTableSize, error) {
	return a.getTableSizes(ctx, a.db)
}

func (a *Activities) getTableSizes(ctx context.Context, db *gorm.DB) ([]app.PSQLTableSize, error) {
	var tables []app.PSQLTableSize

	res := db.WithContext(ctx).
		Find(&tables)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get table sizes: %w", res.Error)
	}

	return tables, nil
}
