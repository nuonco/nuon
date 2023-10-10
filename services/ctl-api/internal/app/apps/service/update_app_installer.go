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
	} `json:"links"`
}

func (c *UpdateAppInstallerRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

//	@BasePath	/v1/apps
//
// Update an app
//
//	@Summary	update an app
//	@Schemes
//	@Description	get an app
//	@Tags			apps
//	@Accept			json
//	@Param			req	body	UpdateAppInstallerRequest	true	"Input"
//	@Produce		json
//	@Param			installer_id	path	string				true	"installer ID"
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		201				{object}	app.AppInstaller
//	@Router			/v1/installers/{installer_id} [PATCH]
func (s *service) UpdateAppInstaller(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req UpdateAppInstallerRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	installer, err := s.updateAppInstaller(ctx, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update app installer: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, installer)
}

func (s *service) updateAppInstaller(ctx context.Context, installerID string, req *UpdateAppInstallerRequest) (*app.AppInstaller, error) {
	currentInstaller := app.AppInstaller{
		ID: installerID,
	}

	installer := app.AppInstaller{
		Metadata: app.AppInstallerMetadata{
			DocumentationURL: req.Links.Documentation,
			GithubURL:        req.Links.Github,
			LogoURL:          req.Links.Logo,
			Description:      req.Description,
			Name:             req.Name,
		},
	}

	res := s.db.WithContext(ctx).Model(&currentInstaller).Updates(installer)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update app installer: %w", res.Error)
	}

	return &installer, nil
}
