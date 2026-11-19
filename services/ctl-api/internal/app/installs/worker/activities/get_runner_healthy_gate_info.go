package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// GetRunnerHealthyGateInfoRequest identifies the install whose runner-healthy
// step generation gate is being evaluated.
type GetRunnerHealthyGateInfoRequest struct {
	InstallID string `validate:"required"`
}

// GetRunnerHealthyGateInfoResponse carries the two inputs step generation needs
// to decide whether to emit the await-runner-healthy step: the install's
// actual runner status, and its desired use_runner setting.
type GetRunnerHealthyGateInfoResponse struct {
	// RunnerStatus is the install's current runner status. Empty if the
	// install has no runner record yet.
	RunnerStatus app.RunnerStatus `json:"runner_status"`
	// UsesRunner is the install's desired use_runner value
	// (InstallConfig.UsesRunner()).
	UsesRunner bool `json:"uses_runner"`
}

// GetRunnerHealthyGateInfo loads the install's desired use_runner setting and
// its runner's actual status, used by installs/workflows/v2 to decide whether
// to generate the await-runner-healthy step for a given phase.
//
// @temporal-gen-v2 activity
// @max-retries 3
func (a *Activities) GetRunnerHealthyGateInfo(ctx context.Context, req GetRunnerHealthyGateInfoRequest) (*GetRunnerHealthyGateInfoResponse, error) {
	install := app.Install{}
	res := a.db.WithContext(ctx).
		Preload("InstallConfig").
		First(&install, "id = ?", req.InstallID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	groupIDs := a.db.WithContext(ctx).
		Model(&app.RunnerGroup{}).
		Select("id").
		Where(app.RunnerGroup{OwnerID: req.InstallID, OwnerType: "installs"})

	var statuses []app.RunnerStatus
	res = a.db.WithContext(ctx).
		Model(&app.Runner{}).
		Scopes(scopes.WithDisableViews).
		Where("runner_group_id IN (?)", groupIDs).
		Order("created_at DESC").
		Limit(1).
		Pluck("status", &statuses)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install runner status: %w", res.Error)
	}

	var status app.RunnerStatus
	if len(statuses) > 0 {
		status = statuses[0]
	}

	return &GetRunnerHealthyGateInfoResponse{
		RunnerStatus: status,
		UsesRunner:   install.InstallConfig.UsesRunner(),
	}, nil
}
