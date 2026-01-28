package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) GetPoliciesConfigByAppConfigID(ctx context.Context, appConfigID string) (*app.AppPoliciesConfig, error) {
	var policiesConfig app.AppPoliciesConfig
	res := h.db.WithContext(ctx).
		Where("app_config_id = ?", appConfigID).
		Preload("Policies").
		Order("created_at DESC").
		First(&policiesConfig)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get policies config")
	}
	return &policiesConfig, nil
}
