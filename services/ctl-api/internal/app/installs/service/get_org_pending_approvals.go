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
// @Security							APIKey
// @Security							OrgID
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

func (s *service) getOrgPendingApprovals(ctx *gin.Context, orgID string) ([]app.WorkflowStepApproval, error) {
	viewName := "install_workflow_step_approvals_pending_v1"
	var approvals []app.WorkflowStepApproval
	res := s.db.WithContext(ctx).
		Omit("contents").
		Scopes(scopes.WithOverrideTable(viewName), scopes.WithOffsetPagination).
		Joins("LEFT JOIN installs ON installs.id = "+viewName+".owner_id AND "+viewName+".owner_type = 'installs'").
		Where("("+viewName+".owner_type != 'installs' OR installs.deleted_at = 0)").
		Where(viewName+".org_id = ?", orgID).
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
