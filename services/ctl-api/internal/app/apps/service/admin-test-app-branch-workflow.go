package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/app-branches/signals"
)

type AdminTestAppBranchWorkflowRequest struct{}

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
	appBranchID := ctx.Param("app_branch_id")
	ab, err := s.getAppBranch(ctx, appBranchID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	fmt.Printf("rb ab.Org.ID", ab.OrgID)

	workflow, err := s.helpers.CreateWorkflow(ctx,
		ab.ID,
		app.WorkflowTypeAppBranchesManualUpdate,
		map[string]string{},
		app.StepErrorBehaviorAbort,
		false,
		&ab.OrgID,
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
