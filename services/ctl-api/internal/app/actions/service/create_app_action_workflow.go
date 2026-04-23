package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/actions/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateAppActionRequest struct {
	Name string `json:"name"`
}

// @ID						CreateAppAction
// @Summary				create an app action
// @Description.markdown	create_app_action_workflow.md
// @Param					app_id	path	string	true	"app ID"
// @Tags					actions
// @Accept					json
// @Param					req	body	CreateAppActionRequest	true	"Input"
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
// @Router					/v1/apps/{app_id}/actions [post]
func (s *service) CreateAppAction(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	_, err = s.findApp(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	var req CreateAppActionWorkflowRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	aw, err := s.createActionWorkflow(ctx, org.ID, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create action workflow: %w", err))
		return
	}

	s.evClient.Send(ctx, aw.ID, &signals.Signal{
		Type: signals.OperationCreated,
	})

	ctx.JSON(http.StatusCreated, aw)
}

type CreateAppActionWorkflowRequest struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels,omitempty"`
}

func (c *CreateAppActionWorkflowRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateAppActionWorkflow
// @Summary				create an app action workflow
// @Description.markdown	create_app_action_workflow.md
// @Param					app_id	path	string	true	"app ID"
// @Tags					actions
// @Accept					json
// @Param					req	body	CreateAppActionWorkflowRequest	true	"Input"
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
// @Router					/v1/apps/{app_id}/action-workflows [post]
func (s *service) CreateAppActionWorkflow(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	_, err = s.findApp(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	var req CreateAppActionWorkflowRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	aw, err := s.createActionWorkflow(ctx, org.ID, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create action workflow: %w", err))
		return
	}

	s.evClient.Send(ctx, aw.ID, &signals.Signal{
		Type: signals.OperationCreated,
	})

	ctx.JSON(http.StatusCreated, aw)
}

func (s *service) createActionWorkflow(ctx *gin.Context, orgID, appID string, req *CreateAppActionWorkflowRequest) (*app.ActionWorkflow, error) {
	return s.actionsHelpers.CreateAction(ctx, &helpers.CreateActionParams{
		AppID:  appID,
		OrgID:  orgID,
		Name:   req.Name,
		Labels: req.Labels,
	})
}
