package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	db "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID									GetOrgPendingApprovals
// @Summary								get all pending workflow step approvals for the org
// @Tags								installs
// @Accept								json
// @Produce								json
// @Security							APIKey && OrgID
// @Param								offset	query	int	false	"offset of results to return"	Default(0)
// @Param								limit	query	int	false	"limit of results to return"	Default(10)
// @Param								page	query	int	false	"page number of results to return"	Default(0)
// @Failure								400	{object}	stderr.ErrResponse
// @Failure								401	{object}	stderr.ErrResponse
// @Failure								403	{object}	stderr.ErrResponse
// @Failure								500	{object}	stderr.ErrResponse
// @Success								200	{array}		app.WorkflowStepApproval
// @Router								/v1/workflows/pending-approvals [GET]
func (s *service) GetOrgPendingApprovals(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	approvals, err := s.getOrgPendingApprovals(ctx, org.ID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get pending approvals"))
		return
	}

	ctx.JSON(http.StatusOK, approvals)
}

// Do not use a view here: the pushed-down org_id makes the planner scan the
// org's full approval history (~2s cold). ANY(ARRAY(...)) fences the active-step
// set so it materializes first and approvals are probed by step id, which keeps
// the plan fast regardless of the planner's choice.
func (s *service) getOrgPendingApprovals(ctx *gin.Context, orgID string) ([]app.WorkflowStepApproval, error) {
	var approvals []app.WorkflowStepApproval
	res := s.db.WithContext(ctx).
		Omit("contents").
		Scopes(scopes.WithOffsetPagination, scopes.ForceReplica).
		Joins("LEFT JOIN installs ON installs.id = install_workflow_step_approvals.owner_id AND install_workflow_step_approvals.owner_type = 'installs'").
		Where("install_workflow_step_approvals.owner_type != 'installs' OR installs.deleted_at = 0").
		Where("install_workflow_step_approvals.deleted_at = 0").
		Where(`install_workflow_step_approvals.install_workflow_step_id = ANY(ARRAY(
			SELECT s.id
			FROM install_workflow_steps s
			JOIN install_workflows w ON w.id = s.install_workflow_id
			LEFT JOIN installs iw ON iw.id = w.owner_id AND w.owner_type = 'installs'
			WHERE w.org_id = ?
			  AND w.finished_at IS NULL
			  AND w.deleted_at = 0
			  AND w.approval_option = 'prompt'
			  AND (w.status->>'status') NOT IN ('cancelled', 'error')
			  AND s.deleted_at = 0
			  AND (s.status->>'status') NOT IN ('auto-skipped', 'cancelled', 'error')
			  AND (w.owner_type != 'installs' OR iw.deleted_at = 0)
		))`, orgID).
		Where("NOT EXISTS (SELECT 1 FROM install_workflow_step_approval_responses r WHERE r.install_workflow_step_approval_id = install_workflow_step_approvals.id AND r.deleted_at = 0)").
		Preload("InstallWorkflowStep").
		Preload("Response").
		Find(&approvals)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get pending approvals")
	}

	approvals, err := db.HandlePaginatedResponse(ctx, approvals)
	if err != nil {
		return nil, errors.Wrap(err, "unable to handle paginated response")
	}

	return approvals, nil
}
