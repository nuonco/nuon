package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type EnsureInstallComponentHealthRequest struct {
	InstallID string `json:"install_id" validate:"required"`
}

// EnsureInstallComponentHealth starts (or confirms already running) this
// install's InstallComponentHealthWorkflow when its runner status is
// not-used. It is the self-healing counterpart to ReconcileUseRunner's own
// start call: ReconcileUseRunner only fires on a status *transition* into
// not-used, so a workflow that failed to start then (or was never confirmed
// running, since that call is best-effort) would otherwise sit without
// health reporting until the next transition. awaitcomponenthealthy calls
// this before every gate wait, so a missing workflow self-heals on the next
// deploy rather than needing an operator to re-trigger ReconcileUseRunner.
//
// A runner that is not not-used (in use, disabled, or never provisioned) is
// a no-op: this activity only ever starts the control-plane stand-in for a
// Runner process, never substitutes for a live one.
//
// @temporal-gen-v2 activity
// @max-retries 2
// @local
func (a *Activities) EnsureInstallComponentHealth(ctx context.Context, req *EnsureInstallComponentHealthRequest) error {
	groupIDs := a.db.WithContext(ctx).
		Model(&app.RunnerGroup{}).
		Select("id").
		Where(app.RunnerGroup{OwnerID: req.InstallID, OwnerType: "installs"})

	var runner app.Runner
	res := a.db.WithContext(ctx).
		Scopes(scopes.WithDisableViews).
		Where("runner_group_id IN (?)", groupIDs).
		Order("created_at DESC").
		First(&runner)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get install runner: %w", res.Error)
	}
	if runner.Status != app.RunnerStatusNotUsed {
		return nil
	}

	if err := a.startInstallComponentHealth(ctx, runner.ID); err != nil {
		return fmt.Errorf("unable to start install component health workflow: %w", err)
	}
	return nil
}
