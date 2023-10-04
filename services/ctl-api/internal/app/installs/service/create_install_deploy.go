package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateInstallDeployRequest struct {
	BuildID string `json:"build_id"`
}

func (c *CreateInstallDeployRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

//	@BasePath	/v1/apps
//
// Deploy a build to an install
//
//	@Summary	deploy a build to an install
//	@Schemes
//	@Description	deploy a build to an install
//	@Param			install_id	path	string						true	"install ID"
//	@Param			req			body	CreateInstallDeployRequest	true	"Input"
//	@Tags			installs
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		201				{object}	app.InstallDeploy
//	@Router			/v1/installs/{install_id}/deploys/ [post]
func (s *service) CreateInstallDeploy(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req CreateInstallDeployRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	deploy, err := s.createInstallDeploy(ctx, installID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	s.hooks.InstallDeployCreated(ctx, installID, deploy.ID)
	ctx.JSON(http.StatusCreated, deploy)
}

func (s *service) createInstallDeploy(ctx context.Context, installID string, req *CreateInstallDeployRequest) (*app.InstallDeploy, error) {
	var build app.ComponentBuild
	res := s.db.WithContext(ctx).
		Preload("ComponentConfigConnection").
		First(&build, "id = ?", req.BuildID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get build %s: %w", req.BuildID, res.Error)
	}

	// create deploy
	installCmp := app.InstallComponent{
		InstallID:   installID,
		ComponentID: build.ComponentConfigConnection.ComponentID,
	}
	deploy := app.InstallDeploy{
		Status:            "queued",
		StatusDescription: "waiting to be deployed to install",
		ComponentBuildID:  req.BuildID,
	}
	err := s.db.WithContext(ctx).First(&installCmp, "install_id = ?", installID).
		Association("InstallDeploys").
		Append(&deploy)
	if err != nil {
		return nil, fmt.Errorf("unable to add install deploy: %w", err)
	}

	return &deploy, nil
}
