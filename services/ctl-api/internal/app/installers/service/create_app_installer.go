package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gosimple/slug"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

type CreateAppInstallerRequest struct {
	AppID       string `validate:"required" json:"app_id"`
	Slug        string `validate:"required" json:"slug"`
	Name        string `validate:"required" json:"name"`
	Description string `validate:"required" json:"description"`

	Links struct {
		Documentation string `validate:"required" json:"documentation"`
		Logo          string `validate:"required" json:"logo"`
		Github        string `validate:"required" json:"github"`
		Homepage      string `validate:"required" json:"homepage"`
		Community     string `validate:"required" json:"community"`
		Demo          string `json:"demo"`
	} `json:"links"`
}

func (c *CreateAppInstallerRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID CreateAppInstaller
// @Summary	create an app installer
// @Description.markdown	create_app_installer.md
// @Tags installers
// @Accept			json
// @Param			req	body	CreateAppInstallerRequest	true	"Input"
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.AppInstaller
// @Router			/v1/installers [POST]
func (s *service) CreateAppInstaller(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateAppInstallerRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	installer, err := s.createAppInstaller(ctx, org.ID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app installer: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, installer)
}

func (s *service) createAppInstaller(ctx context.Context, orgID string, req *CreateAppInstallerRequest) (*app.AppInstaller, error) {
	installer := app.AppInstaller{
		OrgID: orgID,
		AppID: req.AppID,
		Slug:  slug.Make(req.Slug),
		Metadata: app.AppInstallerMetadata{
			Description:      req.Description,
			Name:             req.Name,
			CommunityURL:     req.Links.Community,
			HomepageURL:      req.Links.Homepage,
			DocumentationURL: req.Links.Documentation,
			GithubURL:        req.Links.Github,
			LogoURL:          req.Links.Logo,
			DemoURL:          req.Links.Demo,
		},
	}

	res := s.db.WithContext(ctx).Create(&installer)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app installer: %w", res.Error)
	}

	return &installer, nil
}
