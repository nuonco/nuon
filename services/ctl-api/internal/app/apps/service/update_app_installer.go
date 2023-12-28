package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateAppInstallerRequest struct {
	Name        string `validate:"required" json:"name"`
	Description string `validate:"required" json:"description"`

	Links struct {
		Documentation string `validate:"required" json:"documentation"`
		Logo          string `validate:"required" json:"logo"`
		Github        string `validate:"required" json:"github"`
		Homepage      string `validate:"required" json:"homepage"`
		Community     string `validate:"required" json:"community"`
		Demo          string `validate:"required" json:"demo"`
	} `json:"links"`
}

func (c *UpdateAppInstallerRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @Summary	update an app installer
// @Description.markdown	update_app_installer.md
// @Tags			apps
// @Accept			json
// @Param			req	body	UpdateAppInstallerRequest	true	"Input"
// @Produce		json
// @Param			installer_id	path		string	true	"installer ID"
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.AppInstaller
// @Router			/v1/installers/{installer_id} [PATCH]
func (s *service) UpdateAppInstaller(ctx *gin.Context) {
	installerID := ctx.Param("installer_id")

	var req UpdateAppInstallerRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	installer, err := s.updateAppInstaller(ctx, installerID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update app installer: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, installer)
}

func (s *service) updateAppInstaller(ctx context.Context, installerID string, req *UpdateAppInstallerRequest) (*app.AppInstaller, error) {
	updates := app.AppInstallerMetadata{
		DocumentationURL: req.Links.Documentation,
		GithubURL:        req.Links.Github,
		LogoURL:          req.Links.Logo,
		HomepageURL:      req.Links.Homepage,
		CommunityURL:     req.Links.Community,
		DemoURL:          req.Links.Demo,
		Description:      req.Description,
		Name:             req.Name,
	}

	metadata := app.AppInstallerMetadata{}
	res := s.db.WithContext(ctx).
		Model(&metadata).
		Where("app_installer_id = ?", installerID).
		Updates(updates)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update app installer: %w", res.Error)
	}

	var installer app.AppInstaller
	res = s.db.WithContext(ctx).
		Preload("Metadata").
		Find(&installer, "id = ?", installerID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find installer: %w", res.Error)
	}

	return &installer, nil
}
