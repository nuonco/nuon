package installgrouprelease

import (
	"context"
	"encoding/json"
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
	StatusError           = "error"
)

type Directive struct {
	InstallID      string `json:"install_id"`
	Directive      string `json:"directive"`
	Reason         string `json:"reason,omitempty"`
	WaitingOnRunID string `json:"waiting_on_run_id,omitempty"`
}

func StatusForRelease(reason string) string {
	switch reason {
	case ReasonQueuedBehindCustomer:
		return StatusQueued
	case ReasonRunnerOffline, ReasonRunnerError:
		return StatusError
	default:
		return StatusPendingCustomer
	}
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
		// GORM Update skips the field serializer. A one-element slice is then
		// stored as a JSON object, and the next read fails.
		payload, err := json.Marshal(runs[i].Installs)
		if err != nil {
			return fmt.Errorf("unable to marshal install group run %s installs: %w", runs[i].ID, err)
		}
		if err := db.WithContext(ctx).
			Model(&app.InstallGroupRun{}).
			Where(app.InstallGroupRun{ID: runs[i].ID}).
			Update("installs", string(payload)).Error; err != nil {
			return fmt.Errorf("unable to update install group run %s: %w", runs[i].ID, err)
		}
	}
	return nil
}
