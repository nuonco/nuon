package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type UpdateActionRequest struct {
	Name string `json:"name"`
}

func (c *UpdateActionRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						UpdateAppAction
// @Summary				patch an app action
// @Description.markdown	update_app_action_workflow.md
// @Param					app_id		path	string	true	"app ID"
// @Param					action_id	path	string	true	"action ID"
// @Tags					actions
// @Accept					json
// @Param					req	body	UpdateActionWorkflowRequest	true	"Input"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated  			true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.ActionWorkflow
// @Router					/v1/apps/{app_id}/actions/{action_id} [patch]
func (s *service) UpdateAppAction(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	awID := ctx.Param("action_id")
	_, err = s.findActionWorkflow(ctx, org.ID, awID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", awID, err))
		return
	}

	var req CreateAppActionWorkflowRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	aw, err := s.updateActionWorkflow(ctx, org.ID, awID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, aw)
}

type UpdateActionWorkflowRequest struct {
	Name string `json:"name"`
}

func (c *UpdateActionWorkflowRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						UpdateAppActionWorkflow
// @Summary				patch an app
// @Description.markdown	update_app_action_workflow.md
// @Param					action_workflow_id	path	string	true	"action workflow ID"
// @Tags					actions
// @Accept					json
// @Param					req	body	UpdateActionWorkflowRequest	true	"Input"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated  			true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.ActionWorkflow
// @Router					/v1/action-workflows/{action_workflow_id} [patch]
func (s *service) UpdateActionWorkflow(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	awID := ctx.Param("action_workflow_id")
	_, err = s.findActionWorkflow(ctx, org.ID, awID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", awID, err))
		return
	}

	var req CreateAppActionWorkflowRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	aw, err := s.updateActionWorkflow(ctx, org.ID, awID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, aw)
}

func (s *service) updateActionWorkflow(ctx context.Context, orgID, awID string, req *CreateAppActionWorkflowRequest) (*app.ActionWorkflow, error) {
	aw := app.ActionWorkflow{
		ID:   awID,
		Name: req.Name,
	}

	// up[date where org_id = orgID and id = awID]
	res := s.db.WithContext(ctx).
		Where("org_id = ? AND id = ?", orgID, awID).
		Updates(&aw)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to patch action workflow: %w", res.Error)
	}

	extant, err := s.findActionWorkflow(ctx, orgID, awID)
	if err != nil {
		return nil, err
	}

	return extant, nil
}
