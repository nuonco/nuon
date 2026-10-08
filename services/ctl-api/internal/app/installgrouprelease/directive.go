package installgrouprelease

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const UpdateName = "install-group-directive"

const (
	DirectiveAwait   = "await"
	DirectiveRelease = "release"
)

const (
	ReasonStackPendingCustomer = "stack_pending_customer"
	ReasonRunnerOffline        = "runner_offline"
	ReasonRunnerError          = "runner_error"
	ReasonRunnerDisabled       = "runner_disabled"
	ReasonQueuedBehindCustomer = "queued_behind_customer"
)

const (
	StatusPendingCustomer = "pending_customer"
	StatusQueued          = "queued"
	StatusSuperseded      = "superseded"
)

type Directive struct {
	InstallID      string `json:"install_id"`
	Directive      string `json:"directive"`
	Reason         string `json:"reason,omitempty"`
	WaitingOnRunID string `json:"waiting_on_run_id,omitempty"`
}

func StatusForRelease(reason string) string {
	if reason == ReasonQueuedBehindCustomer {
		return StatusQueued
	}
	return StatusPendingCustomer
}

// SetInstallStatus updates one install inside the group-run rows for a branch run.
// A missing row is not an error: the group may already have finished and dropped it,
// or this install was never recorded.
func SetInstallStatus(ctx context.Context, db *gorm.DB, appBranchRunID, installID string, apply func(*app.InstallGroupRunInstall)) error {
	if appBranchRunID == "" || installID == "" {
		return nil
	}
	var runs []app.InstallGroupRun
	if err := db.WithContext(ctx).
		Where(app.InstallGroupRun{AppBranchRunID: appBranchRunID}).
		Find(&runs).Error; err != nil {
		return fmt.Errorf("unable to list install group runs: %w", err)
	}
	for i := range runs {
		changed := false
		for j := range runs[i].Installs {
			if runs[i].Installs[j].InstallID != installID {
				continue
			}
			apply(&runs[i].Installs[j])
			changed = true
		}
		if !changed {
			continue
		}
		if err := db.WithContext(ctx).
			Model(&app.InstallGroupRun{}).
			Where(app.InstallGroupRun{ID: runs[i].ID}).
			Update("installs", runs[i].Installs).Error; err != nil {
			return fmt.Errorf("unable to update install group run %s: %w", runs[i].ID, err)
		}
	}
	return nil
}
