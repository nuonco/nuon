package activities

import "context"

type AppConfigUsesCustomerManagedRequest struct {
	AppID       string `json:"app_id" validate:"required"`
	AppConfigID string `json:"app_config_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) AppConfigUsesCustomerManaged(ctx context.Context, req AppConfigUsesCustomerManagedRequest) (bool, error) {
	cfg, err := a.loadIntermediateConfig(ctx, req.AppID, req.AppConfigID)
	if err != nil {
		return false, err
	}

	return cfg.CustomerManaged != nil, nil
}
