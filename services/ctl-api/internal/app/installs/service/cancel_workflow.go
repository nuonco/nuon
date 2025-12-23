package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.temporal.io/api/serviceerror"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
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
// @Success						202	{boolean}		true
// @Router						/v1/workflows/{workflow_id}/cancel [post]
func (s *service) CancelWorkflow(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org from context: %w", err))
		return
	}

	workflowID := ctx.Param("workflow_id")

	wf, err := s.getWorkflow(ctx, org.ID, workflowID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get workflow: %w", err))
		return
	}

	if !generics.SliceContains(wf.Status.Status, []app.Status{
		app.StatusInProgress,
		app.StatusPending,
		app.AwaitingApproval,
		app.Status("awaiting-approval"),
	}) {
		s.l.Error("workflow is not cancelable",
			zap.String("workflow_id", wf.ID),
			zap.String("status", string(wf.Status.Status)),
		)
		ctx.Error(stderr.ErrUser{
			Description: "workflow is not cancelable",
			Err:         fmt.Errorf("workflow is not cancelable"),
		})
		return
	}

	if err := s.cancelWorkflow(ctx, wf.ID); err != nil {
		ctx.Error(pkgerrors.Wrap(err, "unable to cancel workflow"))
		return
	}
	if wf.Status.Status == app.StatusPending {
		ctx.JSON(http.StatusAccepted, true)
	}

	id := worker.ExecuteWorkflowIDCallback(signals.RequestSignal{
		EventLoopRequest: eventloop.EventLoopRequest{
			ID: wf.OwnerID,
		},
		Signal: &signals.Signal{
			InstallWorkflowID: wf.ID,
		},
	})
	err = s.evClient.Cancel(ctx, signals.TemporalNamespace, id)
	if err != nil {
		var notFoundErr *serviceerror.NotFound
		if errors.As(err, &notFoundErr) {
			ctx.Error(stderr.ErrNotFound{
				Description: "workflow execution not found in temporal",
				Err:         err,
			})
			return
		}
		ctx.Error(fmt.Errorf("unable to cancel workflow: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, true)
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
// @Success						202	{boolean}		true
// @Router						/v1/install-workflows/{install_workflow_id}/cancel [post]
// @Deprecated
func (s *service) CancelInstallWorkflow(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org from context: %w", err))
		return
	}

	workflowID := ctx.Param("install_workflow_id")

	wf, err := s.getWorkflow(ctx, org.ID, workflowID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install workflow: %w", err))
		return
	}

	if !generics.SliceContains(wf.Status.Status, []app.Status{
		app.StatusInProgress,
		app.StatusPending,
		app.AwaitingApproval,
		app.Status("awaiting-approval"),
	}) {
		s.l.Error("install workflow is not cancelable",
			zap.String("workflow_id", wf.ID),
			zap.String("status", string(wf.Status.Status)),
		)
		ctx.Error(stderr.ErrUser{
			Description: "workflow is not cancelable",
			Err:         fmt.Errorf("workflow is not cancelable"),
		})
		return
	}

	if err := s.cancelWorkflow(ctx, wf.ID); err != nil {
		ctx.Error(pkgerrors.Wrap(err, "unable to cancel workflow"))
		return
	}
	if wf.Status.Status == app.StatusPending {
		ctx.JSON(http.StatusAccepted, true)
	}

	// TODO: cancellation should support workflows by owner type
	id := worker.ExecuteWorkflowIDCallback(signals.RequestSignal{
		EventLoopRequest: eventloop.EventLoopRequest{
			ID: wf.OwnerID,
		},
		Signal: &signals.Signal{
			InstallWorkflowID: wf.ID,
		},
	})

	err = s.evClient.Cancel(ctx, signals.TemporalNamespace, id)
	if err != nil {
		var notFoundErr *serviceerror.NotFound
		if errors.As(err, &notFoundErr) {
			ctx.Error(stderr.ErrNotFound{
				Description: "workflow execution not found in temporal",
				Err:         err,
			})
			return
		}
		ctx.Error(fmt.Errorf("unable to cancel install workflow: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, true)
}

func (s *service) cancelWorkflow(ctx context.Context, installWorkflowID string) error {
	obj := app.Workflow{
		ID: installWorkflowID,
	}

	status := app.NewCompositeStatus(ctx, app.StatusCancelled)
	res := s.db.WithContext(ctx).Model(&obj).Updates(
		map[string]any{
			"status": status,
		})
	if res.Error != nil {
		return pkgerrors.Wrap(res.Error, "unable to update")
	}
	if res.RowsAffected < 1 {
		return pkgerrors.New("no object found to update")
	}

	return nil
}
