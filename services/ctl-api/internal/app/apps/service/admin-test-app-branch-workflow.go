package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/app-branches/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type AdminTestAppBranchWorkflowRequest struct {
	WorkflowType app.WorkflowType `json:"workflow_type" binding:"required"`
}

// @ID						AdminTestAppBranchWorkflow
// @Summary				admin test endpoint to verify triggering an app branch workflow
// @Description.markdown	reprovision_app.md
// @Param					app_branch_id	path	string	true	"app branch ID for your current app"
// @Tags					apps/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminTestAppBranchWorkflowRequest	true	"Input"
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/app-branches/{app_branch_id}/admin-test-app-branch-workflow [POST]
func (s *service) AdminTestAppBranchWorkflow(ctx *gin.Context) {
	var req AdminTestAppBranchWorkflowRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to bind request: %w", err))
		return
	}

	appBranchID := ctx.Param("app_branch_id")
	ab, err := s.getAppBranch(ctx, appBranchID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	org, err := s.getOrg(ctx, ab.OrgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	cctx.SetOrgGinContext(ctx, org)

	workflow, err := s.helpers.CreateWorkflow(ctx,
		ab.ID,
		req.WorkflowType,
		map[string]string{},
		app.StepErrorBehaviorAbort,
		false,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	// TODO: creating an app branch should trigger this signal
	// s.evClient.Send(ctx, req.AppBranchID, &signals.Signal{
	// 	Type: signals.OperationCreated,
	// })

	s.evClient.Send(ctx, ab.ID, &signals.Signal{
		Type:   signals.OperationExecuteFlow,
		FlowID: workflow.ID,
	})

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getOrg(ctx *gin.Context, orgID string) (*app.Org, error) {
	var orgObj *app.Org
	err := s.db.WithContext(ctx).Where("id = ?", orgID).First(&orgObj).Error
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve organization with ID %s: %w", orgID, err)
	}
	return orgObj, nil
}
