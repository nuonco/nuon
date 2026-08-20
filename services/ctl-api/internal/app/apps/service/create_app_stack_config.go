package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/customstacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateAppStackConfigRequest struct {
	Type        app.StackType `json:"type" validate:"required"`
	Description string        `json:"description" validate:"required"`
	Name        string        `json:"name" validate:"required"`

	RunnerNestedTemplateURL string `json:"runner_nested_template_url"`
	VPCNestedTemplateURL    string `json:"vpc_nested_template_url"`

	DeploymentScope string `json:"deployment_scope"`

	CustomNestedStacks []config.CustomNestedStack `json:"custom_nested_stacks"`

	AppConfigID string `json:"app_config_id" validate:"required"`
}

func (c *CreateAppStackConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

func (c *CreateAppStackConfigRequest) toConfig() *config.StackConfig {
	return &config.StackConfig{
		Type:                    string(c.Type),
		Name:                    c.Name,
		Description:             c.Description,
		VPCNestedTemplateURL:    c.VPCNestedTemplateURL,
		RunnerNestedTemplateURL: c.RunnerNestedTemplateURL,
		DeploymentScope:         c.DeploymentScope,
		CustomNestedStacks:      c.CustomNestedStacks,
	}
}

// @ID						CreateAppStackConfig
// @Summary				create an app stack config
// @Description.markdown	create_app_stack_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppStackConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppStackConfig
// @Router					/v1/apps/{app_id}/stack-configs [post]
func (s *service) CreateAppStackConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppStackConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	runnerConfig, err := s.createAppStackConfig(ctx, appID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, runnerConfig)
}

func (s *service) createAppStackConfig(ctx context.Context, appID string, req *CreateAppStackConfigRequest) (*app.AppStackConfig, error) {
	appCloudFormationStackConfig, err := build.StackConfig(req.toConfig(), appID, req.AppConfigID)
	if err != nil {
		return nil, stderr.NewInvalidRequest(err)
	}

	res := s.db.WithContext(ctx).
		Create(appCloudFormationStackConfig)
	if res.Error != nil {
		return nil, res.Error
	}

	// Upload custom nested stack template contents to S3 asynchronously. The
	// activity sets each stack's ContentsHash and marks it ready; consumers gate
	// on Status before generating a stack from these templates.
	if len(appCloudFormationStackConfig.CustomNestedStacks) > 0 {
		q, err := s.queueClient.GetQueueByOwner(ctx, appID, "apps")
		if err != nil {
			return nil, fmt.Errorf("unable to get apps queue for app %s: %w", appID, err)
		}

		if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID: q.ID,
			Signal: &customstacks.Signal{
				AppStackConfigID: appCloudFormationStackConfig.ID,
			},
		}); err != nil {
			return nil, fmt.Errorf("unable to enqueue custom stacks sync signal: %w", err)
		}
	}

	return appCloudFormationStackConfig, nil
}
