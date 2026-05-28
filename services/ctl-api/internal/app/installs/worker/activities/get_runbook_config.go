package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetRunbookConfigByIDRequest struct {
	RunbookConfigID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunbookConfigID
func (a *Activities) GetRunbookConfigByID(ctx context.Context, req GetRunbookConfigByIDRequest) (*app.RunbookConfig, error) {
	var rbConfig app.RunbookConfig
	res := a.db.WithContext(ctx).
		Preload("Steps", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		First(&rbConfig, "id = ?", req.RunbookConfigID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runbook config: %w", res.Error)
	}

	return &rbConfig, nil
}
