package helpers

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) GetComponentAppConfigInTx(ctx context.Context, tx *gorm.DB, appID, branchID string) (*app.AppConfig, error) {
	bound := *h
	bound.db = tx
	if branchID != "" {
		return bound.GetLatestActiveAppConfigForBranch(ctx, appID, branchID)
	}
	return bound.GetLatestActiveAppConfigBare(ctx, appID)
}

func (h *Helpers) GetComponentDependentsInTx(ctx context.Context, tx *gorm.DB, appConfigID, componentID string) ([]string, error) {
	bound := *h
	bound.db = tx
	return bound.GetComponentDependents(ctx, appConfigID, componentID)
}
