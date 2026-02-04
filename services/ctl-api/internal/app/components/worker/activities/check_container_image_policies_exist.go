package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
)

type CheckContainerImagePoliciesExistRequest struct {
	BuildID string `validate:"required"`
}

type CheckContainerImagePoliciesExistResult struct {
	HasPolicies   bool   `json:"has_policies" temporaljson:"has_policies,omitempty"`
	ComponentName string `json:"component_name" temporaljson:"component_name,omitempty"`
	AppConfigID   string `json:"app_config_id" temporaljson:"app_config_id,omitempty"`
}

// @temporal-gen activity
// @max-retries 1
// @schedule-to-close-timeout 1m
// @start-to-close-timeout 30s
func (a *Activities) CheckContainerImagePoliciesExist(ctx context.Context, req *CheckContainerImagePoliciesExistRequest) (*CheckContainerImagePoliciesExistResult, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("build_id", req.BuildID))

	l.Info("checking if container image policies exist")

	build, err := a.getBuildWithAppConfig(ctx, req.BuildID)
	if err != nil {
		l.Error("unable to get build with app config", zap.Error(err))
		return nil, errors.Wrap(err, "unable to get build with app config")
	}

	componentName := build.ComponentConfigConnection.Component.Name

	appConfigs := build.ComponentConfigConnection.Component.App.AppConfigs
	if len(appConfigs) == 0 {
		l.Info("no app config found, no policies to evaluate")
		return &CheckContainerImagePoliciesExistResult{
			HasPolicies:   false,
			ComponentName: componentName,
		}, nil
	}
	appConfigID := appConfigs[0].ID

	l = l.With(zap.String("app_config_id", appConfigID))

	policiesConfig, err := a.appsHelpers.GetPoliciesConfigByAppConfigID(ctx, appConfigID)
	if err != nil {
		l.Info("no policies config found, no policies to evaluate")
		return &CheckContainerImagePoliciesExistResult{
			HasPolicies:   false,
			ComponentName: componentName,
			AppConfigID:   appConfigID,
		}, nil
	}

	applicablePolicies := a.filterContainerImagePolicies(policiesConfig.Policies, componentName)

	l.Info("checked container image policies", zap.Int("count", len(applicablePolicies)))

	return &CheckContainerImagePoliciesExistResult{
		HasPolicies:   len(applicablePolicies) > 0,
		ComponentName: componentName,
		AppConfigID:   appConfigID,
	}, nil
}
