package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type PlannedComponentBuild struct {
	ComponentID                 string `json:"component_id"`
	BuildID                     string `json:"build_id"`
	ComponentConfigConnectionID string `json:"component_config_connection_id"`
	DeployID                    string `json:"deploy_id"`
	WaitForBuild                bool   `json:"wait_for_build"`
}

type InstallPreflightRequest struct {
	FlowID                 string                  `json:"flow_id" validate:"required"`
	InstallID              string                  `json:"install_id" validate:"required"`
	DesiredAppConfigID     string                  `json:"desired_app_config_id"`
	CheckStackOutdated     bool                    `json:"check_stack_outdated"`
	PlannedComponentBuilds []PlannedComponentBuild `json:"planned_component_builds"`
}

type InstallPreflightResponse struct {
	Findings []*compositeerrors.CompositeErrorData `json:"findings"`
}

type installPreflightCheck struct {
	name string
	run  func(context.Context, InstallPreflightRequest) ([]*compositeerrors.CompositeErrorData, error)
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) InstallPreflight(ctx context.Context, req InstallPreflightRequest) (*InstallPreflightResponse, error) {
	l := temporalzap.GetActivityLogger(ctx).With(
		zap.String("flow_id", req.FlowID),
		zap.String("install_id", req.InstallID),
	)
	l.Info("starting install preflight checks")

	result, err := a.runInstallPreflightChecks(ctx, req)
	if err != nil {
		l.Error("install preflight check failed", zap.Error(err))
		return nil, err
	}

	l.Info("completed install preflight checks", zap.Int("finding_count", len(result.Findings)))
	return result, nil
}

func (a *Activities) runInstallPreflightChecks(ctx context.Context, req InstallPreflightRequest) (*InstallPreflightResponse, error) {
	checks := []installPreflightCheck{
		{name: "install-stack-outdated", run: a.checkInstallStackOutdated},
		{name: "component-builds", run: a.checkComponentBuilds},
	}

	findings := make([]*compositeerrors.CompositeErrorData, 0)
	for _, check := range checks {
		current, err := check.run(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("preflight check %q failed: %w", check.name, err)
		}
		findings = append(findings, current...)
	}

	return &InstallPreflightResponse{Findings: findings}, nil
}
