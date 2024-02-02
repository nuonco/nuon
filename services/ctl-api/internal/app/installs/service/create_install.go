package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

type CreateInstallRequest struct {
	Name string `json:"name" validate:"required"`

	AWSAccount struct {
		Region     string `json:"region"`
		IAMRoleARN string `json:"iam_role_arn" validate:"required"`
	} `json:"aws_account" validate:"required"`

	Inputs map[string]*string `json:"inputs" validate:"required"`
}

func (c *CreateInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID CreateInstall
// @Summary	create an app install
// @Description.markdown	create_install.md
// @Param			app_id	path	string					true	"app ID"
// @Param			req		body	CreateInstallRequest	true	"Input"
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.Install
// @Router			/v1/apps/{app_id}/installs [post]
func (s *service) CreateInstall(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")

	var req CreateInstallRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	install, err := s.createInstall(ctx, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	s.hooks.Created(ctx, install.ID, org.SandboxMode)
	ctx.JSON(http.StatusCreated, install)
}

func (s *service) createInstall(ctx context.Context, appID string, req *CreateInstallRequest) (*app.Install, error) {
	parentApp := app.App{}
	res := s.db.WithContext(ctx).Preload("Components").
		Preload("AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC")
		}).
		Preload("AppRunnerConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_runner_configs.created_at DESC")
		}).
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}
	if len(parentApp.AppSandboxConfigs) < 1 {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("app does not have any sandbox configs"),
			Description: "please make create at least one app sandbox config first",
		}
	}
	if len(parentApp.AppRunnerConfigs) < 1 {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("app does not have any runner configs"),
			Description: "please make create at least one app runner config first",
		}
	}

	if err := s.validateInstallInputs(ctx, appID, req.Inputs); err != nil {
		return nil, err
	}

	install := app.Install{
		AppID:             appID,
		Name:              req.Name,
		Status:            "queued",
		StatusDescription: "waiting for event loop to start and provision install",
		AWSAccount: app.AWSAccount{
			Region:     req.AWSAccount.Region,
			IAMRoleARN: req.AWSAccount.IAMRoleARN,
		},
		InstallInputs: []app.InstallInputs{
			{
				Values: req.Inputs,
			},
		},
		AppSandboxConfigID: parentApp.AppSandboxConfigs[0].ID,
		AppRunnerConfigID:  parentApp.AppRunnerConfigs[0].ID,
	}

	res = s.db.WithContext(ctx).Create(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install: %w", res.Error)
	}

	if err := s.componentHelpers.EnsureInstallComponents(ctx, appID, []string{install.ID}); err != nil {
		return nil, fmt.Errorf("unable to ensure install components: %w", err)
	}

	return &install, nil
}
