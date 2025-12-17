package workflow

import (
	"errors"
	"strconv"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	pkgErrors "github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/poll"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type WaitForApprovalResponseRequest struct {
	WorkflowID string `json:"workflow_id" validate:"required"`
	StepID     string `json:"step_id" validate:"required"`
	MaxTS      *time.Time
}

// @temporal-gen workflow
// @execution-timeout 720h
// @task-timeout 1m
// @id-template {{.CallerID}}-{{.Req.StepID}}-wait-for-approval-response
func (w *Workflows) WaitForApprovalResponse(ctx workflow.Context, req *WaitForApprovalResponseRequest) (*app.WorkflowStepApprovalResponse, error) {
	maxTS := workflow.Now(ctx).Add(time.Hour * 24 * 30)
	if req.MaxTS != nil {
		maxTS = *req.MaxTS
	}

	resp, err := w.waitForApprovalResponse(ctx, req.WorkflowID, req.StepID, maxTS)
	if err != nil {
		if errors.Is(err, poll.ContinueAsNewErr) {
			if req.MaxTS == nil {
				req.MaxTS = &maxTS
			}

			return nil, workflow.NewContinueAsNewError(ctx, w.WaitForApprovalResponse, req)
		}
		return nil, err
	}

	return resp, nil
}

func (w *Workflows) waitForApprovalResponse(ctx workflow.Context, workflowID, stepID string, maxTS time.Time) (*app.WorkflowStepApprovalResponse, error) {
	if err := poll.Poll(ctx, w.v, poll.PollOpts{
		MaxTS:                      maxTS,
		InitialInterval:            time.Second * 15,
		MaxInterval:                time.Minute * 15,
		BackoffFactor:              1,
		ContinueAsNewAfterAttempts: 60,
		PostAttemptHook: func(ctx workflow.Context, dur time.Duration) error {
			l, err := log.WorkflowLogger(ctx)
			if err != nil {
				return pkgErrors.Wrap(err, "unable to get workflow logger")
			}

			l.Debug("checking approval status again in "+dur.String(), zap.Duration("duration", dur))
			return nil
		},
		Fn: func(ctx workflow.Context) error {
			stp, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, stepID)
			if err != nil {
				return pkgErrors.Wrap(err, "unable to get flow step")
			}

			if stp.Approval == nil {
				return pkgErrors.New("Approval does not exist yet")
			}

			// get latest workflow to ensure we have the latest state since approval options can change
			latestFlw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, workflowID)
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
							"step_idx": stp.Idx,
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
		return nil, err
	}

	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, stepID)
	if err != nil {
		return nil, pkgErrors.Wrap(err, "unable to get approval step")
	}

	// should never happen due to polling above, but for sanity.
	if step.Approval.Response == nil {
		return nil, pkgErrors.New("approval response is still nil after polling")
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
