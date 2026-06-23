package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	pkgerrors "github.com/pkg/errors"
)

// @ID							CancelWorkflow
// @Summary						cancel an ongoing workflow
// @Description.markdown		cancel_workflow.md
// @Param workflow_id	path	string true "workflow ID"
// @Tags						installs
// @Accept						json
// @Produce						json
// @Security					APIKey
// @Security					OrgID
// @Failure						400	{object}	stderr.ErrResponse
// @Failure						401	{object}	stderr.ErrResponse
// @Failure						403	{object}	stderr.ErrResponse
// @Failure						404	{object}	stderr.ErrResponse
// @Failure						500	{object}	stderr.ErrResponse
// @Success				202	{object}	app.EmptyResponse
// @Router						/v1/workflows/{workflow_id}/cancel [post]
func (s *service) CancelWorkflow(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org from context: %w", err))
		return
	}

	workflowID := ctx.Param("workflow_id")
	if err := s.cancelSingleWorkflow(ctx, org.ID, workflowID); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusAccepted, app.EmptyResponse{})
}

// TODO: Remove. Deprecated.
// @ID							CancelInstallWorkflow
// @Summary						cancel an ongoing install workflow
// @Description.markdown		cancel_workflow.md
// @Param install_workflow_id	path	string true "install workflow ID"
// @Tags						installs
// @Accept						json
// @Produce						json
// @Security					APIKey
// @Security					OrgID
// @Failure						400	{object}	stderr.ErrResponse
// @Failure						401	{object}	stderr.ErrResponse
// @Failure						403	{object}	stderr.ErrResponse
// @Failure						404	{object}	stderr.ErrResponse
// @Failure						500	{object}	stderr.ErrResponse
// @Success				202	{object}	app.EmptyResponse
// @Router						/v1/install-workflows/{install_workflow_id}/cancel [post]
// @Deprecated
func (s *service) CancelInstallWorkflow(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org from context: %w", err))
		return
	}

	workflowID := ctx.Param("install_workflow_id")
	if err := s.cancelSingleWorkflow(ctx, org.ID, workflowID); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusAccepted, app.EmptyResponse{})
}

// cancelSingleWorkflow validates, cancels, and signals a single workflow.
// It returns an error suitable for ctx.Error().
func (s *service) cancelSingleWorkflow(ctx *gin.Context, orgID, workflowID string) error {
	wf, err := s.getWorkflow(ctx, orgID, workflowID)
	if err != nil {
		return fmt.Errorf("unable to get workflow: %w", err)
	}

	if !generics.SliceContains(wf.Status.Status, []app.Status{
		app.StatusInProgress,
		app.StatusPending,
		app.AwaitingApproval,
		app.Status("awaiting-approval"),
		app.StatusFailedPendingRetry,
	}) {
		return fmt.Errorf("workflow is not cancelable (status: %s)", wf.Status.Status)
	}

	// If the workflow hasn't started yet, cancel it directly in the DB —
	// there is no signal to cancel.
	if wf.Status.Status == app.StatusPending {
		if err := s.cancelWorkflow(ctx, wf.ID); err != nil {
			return fmt.Errorf("unable to cancel workflow: %w", err)
		}
		return nil
	}

	if _, err := s.flowsClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{
		InstallWorkflowID: wf.ID,
	}); err != nil {
		// Orphaned workflows (e.g. weeks-old rows stuck in-progress) no longer
		// have a live execute-flow queue signal — it was soft-deleted once the
		// handler finished. There is nothing to signal, so flip the row
		// directly in the DB rather than leaving it stuck in-progress forever.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if dbErr := s.cancelWorkflow(ctx, wf.ID); dbErr != nil {
				return fmt.Errorf("unable to cancel orphaned workflow: %w", dbErr)
			}
			return nil
		}
		return fmt.Errorf("unable to cancel workflow via queues: %w", err)
	}

	return nil
}

// findCancelableStep returns the first in-progress or awaiting-approval step, if any.
func (s *service) findCancelableStep(wf *app.Workflow) *app.WorkflowStep {
	for i := range wf.Steps {
		switch wf.Steps[i].Status.Status {
		case app.StatusInProgress, app.AwaitingApproval, app.Status("awaiting-approval"), app.StatusFailedPendingRetry:
			return &wf.Steps[i]
		}
	}
	return nil
}

func (s *service) cancelWorkflow(ctx context.Context, installWorkflowID string) error {
	// Load-then-Save so Workflow.BeforeSave sees Type/Metadata and can
	// recompute name to the past-tense title.
	var obj app.Workflow
	if err := s.db.WithContext(ctx).Where("id = ?", installWorkflowID).Take(&obj).Error; err != nil {
		return pkgerrors.Wrap(err, "unable to load workflow")
	}

	obj.Status = app.NewCompositeStatus(ctx, app.StatusCancelled)
	obj.FinishedAt = time.Now()

	res := s.db.WithContext(ctx).Save(&obj)
	if res.Error != nil {
		return pkgerrors.Wrap(res.Error, "unable to update")
	}
	if res.RowsAffected < 1 {
		return pkgerrors.New("no object found to update")
	}

	return nil
}
