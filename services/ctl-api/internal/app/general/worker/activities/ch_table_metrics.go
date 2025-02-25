package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetCHTableMetricsRequest struct{}

// @temporal-gen activity
func (a *Activities) GetCHTableMetrics(ctx context.Context, req GetCHTableMetricsRequest) ([]app.CHTableSize, error) {
	return a.getCHTableSizes(ctx, a.chDB)
}

func (a *Activities) getCHTableSizes(ctx context.Context, db *gorm.DB) ([]app.CHTableSize, error) {
	var tables []app.CHTableSize

	res := db.WithContext(ctx).
		Find(&tables)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get table sizes: %w", res.Error)
	}

	return tables, nil
}
