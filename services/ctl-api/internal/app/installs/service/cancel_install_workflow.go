package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

// @ID						CancelInstallWorkflow
// @Summary				cancel an ongoing install workflow
// @Description.markdown	cancel_install_workflow.md
// @Param install_workflow_id	path	string true "install workflow ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				202	{boolean}		true
// @Router					/v1/install-workflows/{install_workflow_id}/cancel [post]
func (s *service) CancelInstallWorkflow(ctx *gin.Context) {
	workflowID := ctx.Param("install_workflow_id")

	wf, err := s.getInstallWorkflow(ctx, workflowID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install workflow: %w", err))
		return
	}

	if !generics.SliceContains(wf.Status.Status, []app.Status{
		app.StatusInProgress,
		app.StatusPending,
	}) {
		ctx.Error(fmt.Errorf("install workflow is not cancelable"))
		return
	}

	if wf.Status.Status == app.StatusPending {
		if err := s.cancelInstallWorkflow(ctx, wf.ID); err != nil {
			ctx.Error(errors.Wrap(err, "unable to cancel workflow"))
			return
		}
	}

	if wf.Status.Status == app.StatusInProgress {
		id := fmt.Sprintf("sig-execute-workflow-%s", wf.InstallID)
		err = s.evClient.Cancel(ctx, signals.TemporalNamespace, id)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to cancel install workflow: %w", err))
			return
		}
	}

	ctx.JSON(http.StatusAccepted, true)
}

func (s *service) cancelInstallWorkflow(ctx context.Context, installWorkflowID string) error {
	obj := app.InstallWorkflow{
		ID: installWorkflowID,
	}

	status := app.NewCompositeStatus(ctx, app.StatusCancelled)
	res := s.db.WithContext(ctx).Model(obj).Updates(
		map[string]any{
			"status": status,
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update")
	}
	if res.RowsAffected < 1 {
		return errors.New("no object found to update")
	}

	return nil
}
