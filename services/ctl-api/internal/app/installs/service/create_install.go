package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

type CreateInstallRequest struct {
	Name string `json:"name" validate:"required"`

	AWSAccount struct {
		Region     string `json:"region" validate:"oneof=us-east-1 us-east-2 us-west-2"`
		IAMRoleARN string `json:"iam_role_arn" validate:"required"`
	} `json:"aws_account" validate:"required"`
}

func (c *CreateInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

//	@BasePath	/v1/apps
//
// Create an app install
//
//	@Summary	create an app install
//	@Schemes
//	@Description	create an app install
//	@Param			app_id	path	string					true	"app ID"
//	@Param			req		body	CreateInstallRequest	true	"Input"
//	@Tags			installs
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		201				{object}	app.Install
//	@Router			/v1/apps/{app_id}/installs/ [post]
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
	res := s.db.WithContext(ctx).Preload("Components").Preload("SandboxRelease").First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	installCmps := make([]app.InstallComponent, 0)
	for _, cmp := range parentApp.Components {
		installCmps = append(installCmps, app.InstallComponent{
			ComponentID: cmp.ID,
		})
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
		SandboxReleaseID:  parentApp.SandboxRelease.ID,
		InstallComponents: installCmps,
	}

	res = s.db.WithContext(ctx).Create(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install: %w", res.Error)
	}
	return &install, nil
}
