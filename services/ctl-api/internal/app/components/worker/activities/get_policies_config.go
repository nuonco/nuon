package activities

import (
	"context"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"go.uber.org/zap"
)

type GetPoliciesConfigByAppConfigIDRequest struct {
	AppConfigID string `validate:"required"`
}

// @temporal-gen activity
// @max-retries 1
func (a *Activities) GetPoliciesConfigByAppConfigID(ctx context.Context, req *GetPoliciesConfigByAppConfigIDRequest) (*app.AppPoliciesConfig, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("app_config_id", req.AppConfigID))

	policiesConfig, err := a.appsHelpers.GetPoliciesConfigByAppConfigID(ctx, req.AppConfigID)
	if err != nil {
		l.Info("no policies config found", zap.Error(err))
		return nil, nil
	}

	return policiesConfig, nil
}
