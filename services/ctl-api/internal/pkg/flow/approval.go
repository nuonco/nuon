package flow

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	pkgErrors "github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/poll"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func (c *WorkflowConductor[DomainSignal]) waitForApprovalResponse(ctx workflow.Context, flw *app.Workflow, step *app.WorkflowStep, stepIdx int) (*app.WorkflowStepApprovalResponse, error) {
	if err := poll.Poll(ctx, c.V, poll.PollOpts{
		MaxTS:           workflow.Now(ctx).Add(time.Hour * 24 * 30),
		InitialInterval: time.Second * 15,
		MaxInterval:     time.Minute * 15,
		BackoffFactor:   1,
		PostAttemptHook: func(ctx workflow.Context, dur time.Duration) error {
			l, err := log.WorkflowLogger(ctx)
			if err != nil {
				return pkgErrors.Wrap(err, "unable to get workflow logger")
			}

			l.Debug("checking approval status again in "+dur.String(), zap.Duration("duration", dur))
			return nil
		},
		Fn: func(ctx workflow.Context) error {
			stp, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
			if err != nil {
				return pkgErrors.Wrap(err, "unable to get flow step")
			}

			if stp.Approval == nil {
				return pkgErrors.New("Approval does not exist yet")
			}

			// get latest workflow to ensure we have the latest state since approval options can change
			latestFlw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, flw.ID)
			if err != nil {
				return errors.Join(pkgErrors.Wrap(err, "unable to get latest flow"), poll.NonRetryableError)
			}

			if latestFlw.ApprovalOption == app.InstallApprovalOptionApproveAll {
				if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
					ID: latestFlw.ID,
					Status: app.CompositeStatus{
						Status:                 app.WorkflowStepApprovalStatusApproved,
						StatusHumanDescription: "auto approved for step " + strconv.Itoa(stp.Idx+1),
						Metadata: map[string]any{
							"step_idx": step.Idx,
							"status":   "auto-approved",
						},
					},
				}); err != nil {
					return errors.Join(pkgErrors.Wrap(err, "unable to update flow status"), poll.NonRetryableError)
				}

				_, err := activities.AwaitCreateApprovalResponse(ctx, activities.CreateStepApprovalResponseRequest{
					StepApprovalID: stp.Approval.ID,
					Type:           app.WorkflowStepApprovalResponseTypeApprove,
					Note:           "auto-approved",
				})
				if err != nil {
					return errors.Join(pkgErrors.Wrap(err, "unable to create auto-approval response"), poll.NonRetryableError)
				}

				return nil
			}

			if stp.Approval.Response == nil {
				return pkgErrors.New("approval does not yet have a response")
			}

			return nil
		},
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: step.ID,
				Status: app.NewCompositeTemporalStatus(ctx, app.WorkflowStepApprovalStatusApprovalExpired, map[string]any{
					"err_message": "approval was not accepted",
				}),
			})

			return nil, c.handleCancellation(ctx, err, step.ID, stepIdx, flw)
		}
	}

	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
	if err != nil {
		return nil, pkgErrors.Wrap(err, "unable to get approval step")
	}

	if step.Approval.Response.Type == app.WorkflowStepApprovalResponseTypeDeny {
		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
			StepID: step.ID,
			Status: app.WorkflowStepApprovalStatusApprovalDenied,
		}); err != nil {
			return nil, pkgErrors.Wrap(err, "unable to update step target status")
		}
	}

	return step.Approval.Response, nil
}
