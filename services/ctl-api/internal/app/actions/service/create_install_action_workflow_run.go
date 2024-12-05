package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type CreateInstallActionWorkflowRunRequest struct {
	ActionWorkFlowConfigID string `json:"action_workflow_config_id" binding:"required"`
}

func (c *CreateInstallActionWorkflowRunRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID CreateInstallActionWorkflowRun
// @Summary	create an action workflow run for an install
// @Description.markdown	create_install_action_workflow_run.md
// @Tags			actions
// @Accept			json
// @Param			install_id	path	string	true	"install ID"
// @Param			req	body CreateInstallActionWorkflowRunRequest true	"Input"
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}  app.InstallActionWorkflowRun
// @Router		/v1/installs/{install_id}/action-workflows/runs [post]
func (s *service) CreateInstallActionWorkflowRun(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	var req CreateInstallActionWorkflowRunRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	_, err = s.findActionWorkflowConfig(ctx, org.ID, req.ActionWorkFlowConfigID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	run, err := s.createActionWorkflowRun(ctx, org.ID, installID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	// TODO: trigger a signal the action workflow run

	ctx.JSON(http.StatusCreated, run)
}

func (s *service) createActionWorkflowRun(ctx *gin.Context, orgID, installID string, req CreateInstallActionWorkflowRunRequest) (*app.InstallActionWorkflowRun, error) {
	newRun := app.InstallActionWorkflowRun{
		OrgID:                  orgID,
		InstallID:              installID,
		ActionWorkflowConfigID: req.ActionWorkFlowConfigID,
		Status:                 app.InstallActionRunStatusQueued,
		StatusDescription:      "Queued",
	}

	res := s.db.WithContext(ctx).
		Create(&newRun)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create action workflow: %w", res.Error)
	}

	return nil, nil
}
