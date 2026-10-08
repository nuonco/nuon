package activities

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

var installGroupSkipSignalTypes = map[signal.SignalType]struct{}{
	"app-branch-update-install-group": {},
	"app-branch-post-deploy-runbooks": {},
}

var installGroupSkipStatuses = map[app.Status]struct{}{
	app.StatusPending:      {},
	app.StatusNotAttempted: {},
	app.StatusQueued:       {},
}

type SkipPendingInstallGroupFollowOnStepsInput struct {
	InstallWorkflowID string `json:"install_workflow_id" validate:"required"`
	InstallGroupID    string `json:"install_group_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) SkipPendingInstallGroupFollowOnSteps(ctx context.Context, input *SkipPendingInstallGroupFollowOnStepsInput) error {
	skipStatus := app.CompositeStatus{
		Status:                 app.StatusUserSkipped,
		StatusHumanDescription: "install group plan skipped, deploy skipped",
	}

	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var steps []app.WorkflowStep
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.WorkflowStep{
				InstallWorkflowID: input.InstallWorkflowID,
				ExecutionType:     app.WorkflowStepExecutionTypeSystem,
			}).
			Find(&steps).Error; err != nil {
			return errors.Wrap(err, "unable to query install group follow-on steps")
		}

		for _, step := range steps {
			if !shouldSkipInstallGroupFollowOn(step, input.InstallGroupID) {
				continue
			}
			next, err := statusactivities.NextCompositeStatus(ctx, step.Status, skipStatus)
			if err != nil {
				return err
			}
			res := tx.Model(&app.WorkflowStep{ID: step.ID}).Updates(map[string]any{
				"status": next,
			})
			if res.Error != nil {
				return errors.Wrap(res.Error, "unable to update flow step")
			}
			if res.RowsAffected < 1 {
				return errors.New("no object found to update")
			}
		}
		return nil
	})
}

func shouldSkipInstallGroupFollowOn(step app.WorkflowStep, installGroupID string) bool {
	if _, ok := installGroupSkipStatuses[step.Status.Status]; !ok {
		return false
	}
	if step.QueueSignal == nil || step.QueueSignal.Signal == nil {
		return false
	}
	sig := step.QueueSignal.Signal
	if _, ok := installGroupSkipSignalTypes[sig.Type()]; !ok {
		return false
	}
	raw, err := json.Marshal(sig)
	if err != nil {
		return false
	}
	var payload struct {
		InstallGroupID string `json:"install_group_id"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return false
	}
	return payload.InstallGroupID == installGroupID
}
