package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type ReconcileInstallRequest struct {
	ExpectedAppConfigID string `json:"expected_app_config_id" validate:"required"`
	PlanOnly            bool   `json:"plan_only"`
}

type ReconcileInstallResponse struct {
	WorkflowID                string                         `json:"workflow_id,omitempty"`
	InstallAppConfigVersionID string                         `json:"install_app_config_version_id,omitempty"`
	Reconciliation            *InstallReconciliationResponse `json:"reconciliation"`
}

// @ID						ReconcileInstall
// @Summary				reconcile an install onto its app branch
// @Description			Starts the workflow that brings the install onto its app branch's latest deployable config. expected_app_config_id is the config the caller reviewed; the request fails with 409 when the branch has since moved to another config, when the install is blocked, or when it has no deployable target. When the install already runs the target config, or a workflow is already rolling it out, nothing is created and the current state is returned with 200. plan_only runs the workflow without applying it. Requires the install-reconciliation org feature.
// @Param					install_id	path	string					true	"install ID"
// @Param					req			body	ReconcileInstallRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	service.ReconcileInstallResponse
// @Success				201	{object}	service.ReconcileInstallResponse
// @Router					/v1/installs/{install_id}/reconcile [post]
func (s *service) ReconcileInstall(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireInstallReconciliation(ctx, org.ID); err != nil {
		ctx.Error(err)
		return
	}

	var req ReconcileInstallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := s.v.Struct(req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	install, err := s.getReconcilableInstall(ctx, org.ID, ctx.Param("install_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	target, err := s.helpers.ResolveReconcileTarget(ctx, install)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to resolve reconciliation target: %w", err))
		return
	}

	if !target.HasTarget() {
		ctx.Error(reconcileBlockedError(target))
		return
	}
	if req.ExpectedAppConfigID != target.AppConfigID {
		ctx.Error(stderr.ErrConflict{
			Err:         fmt.Errorf("app branch moved from %s to %s since the reconcile was reviewed", req.ExpectedAppConfigID, target.AppConfigID),
			Description: "The app branch has changed since you reviewed this update. Review the latest changes and try again.",
		})
		return
	}

	switch target.Status() {
	case installhelpers.ReconcileStatusCurrent, installhelpers.ReconcileStatusInProgress:
		resp, err := s.installReconciliation(ctx, install, target)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to get install reconciliation: %w", err))
			return
		}
		ctx.JSON(http.StatusOK, ReconcileInstallResponse{
			WorkflowID:     target.InFlightWorkflowID,
			Reconciliation: resp,
		})
		return
	case installhelpers.ReconcileStatusBlocked:
		ctx.Error(reconcileBlockedError(target))
		return
	}

	update, err := s.helpers.CreateAppBranchConfigUpdateWorkflow(ctx, installhelpers.AppBranchConfigUpdateInput{
		InstallID:      install.ID,
		NewAppConfigID: target.AppConfigID,
		AppBranchRunID: target.AppBranchRunID,
		InstallGroupID: target.InstallGroupID,
		PlanOnly:       req.PlanOnly,
		TriggeredBy:    app.WorkflowTriggeredByReconcile,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create reconcile workflow: %w", err))
		return
	}

	target.InFlightWorkflowID = update.WorkflowID
	resp, err := s.installReconciliation(ctx, install, target)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install reconciliation: %w", err))
		return
	}
	ctx.JSON(http.StatusCreated, ReconcileInstallResponse{
		WorkflowID:                update.WorkflowID,
		InstallAppConfigVersionID: update.InstallAppConfigVersionID,
		Reconciliation:            resp,
	})
}

func reconcileBlockedError(target *installhelpers.ReconcileTarget) error {
	if len(target.Blockers) == 0 {
		return stderr.ErrConflict{
			Err:         fmt.Errorf("install cannot be reconciled"),
			Description: "This install cannot be reconciled right now.",
		}
	}
	blocker := target.Blockers[0]
	return stderr.ErrConflict{
		Err:         fmt.Errorf("install cannot be reconciled: %s", blocker.Code),
		Description: blocker.Message,
	}
}
