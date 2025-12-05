package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
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
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	app, err := s.createActionWorkflow(ctx, org.ID, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create action workflow: %w", err))
		return
	}

	if err := s.actionsHelpers.EnsureInstallAction(ctx, appID, nil); err != nil {
		ctx.Error(errors.Wrap(err, "unable to ensure actions"))
		return
	}

	s.evClient.Send(ctx, app.ID, &signals.Signal{
		Type: signals.OperationCreated,
	})

	ctx.JSON(http.StatusCreated, app)
}

type CreateAppActionWorkflowRequest struct {
	Name string `json:"name"`
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
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	app, err := s.createActionWorkflow(ctx, org.ID, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create action workflow: %w", err))
		return
	}

	if err := s.actionsHelpers.EnsureInstallAction(ctx, appID, nil); err != nil {
		ctx.Error(errors.Wrap(err, "unable to ensure actions"))
		return
	}

	s.evClient.Send(ctx, app.ID, &signals.Signal{
		Type: signals.OperationCreated,
	})

	ctx.JSON(http.StatusCreated, app)
}

func (s *service) createActionWorkflow(ctx context.Context, orgID, appID string, req *CreateAppActionWorkflowRequest) (*app.ActionWorkflow, error) {
	newAW := app.ActionWorkflow{
		AppID:  appID,
		OrgID:  orgID,
		Name:   req.Name,
		Status: app.ActionWorkflowStatusActive,
	}

	res := s.db.WithContext(ctx).
		Create(&newAW)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create action workflow: %w", res.Error)
	}

	return &newAW, nil
}
