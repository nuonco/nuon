package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateInstallerRequest struct {
	AppIDs []string `validate:"required" json:"app_ids"`
	Name   string   `validate:"required" json:"name"`

	Metadata struct {
		Description      string `validate:"required" json:"description"`
		DocumentationURL string `validate:"required" json:"documentation_url"`
		LogoURL          string `validate:"required" json:"logo_url"`
		GithubURL        string `validate:"required" json:"github_url"`
		HomepageURL      string `validate:"required" json:"homepage_url"`
		CommunityURL     string `validate:"required" json:"community_url"`
		FaviconURL       string `validate:"required" json:"favicon_url" `

		OgImageURL          string `json:"og_image_url" `
		DemoURL             string `json:"demo_url"`
		PostInstallMarkdown string `json:"post_install_markdown"`
		FooterMarkdown      string `json:"footer_markdown"`
		CopyrightMarkdown   string `json:"copyright_markdown"`
	} `json:"metadata"`
}

func (c *CreateInstallerRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

//	@ID						CreateInstaller
//	@Summary				create an installer
//	@Description.markdown	create_installer.md
//	@Tags					installers
//	@Accept					json
//	@Param					req	body	CreateInstallerRequest	true	"Input"
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				201	{object}	app.Installer
//	@Router					/v1/installers [POST]
func (s *service) CreateInstaller(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateInstallerRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	installer, err := s.createInstaller(ctx, org.ID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create installer: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, installer)
}

func (s *service) createInstaller(ctx context.Context, orgID string, req *CreateInstallerRequest) (*app.Installer, error) {
	apps := make([]app.App, 0)
	for _, appID := range req.AppIDs {
		apps = append(apps, app.App{ID: appID})
	}

	installer := app.Installer{
		OrgID: orgID,
		Apps:  apps,
		Type:  app.InstallerTypeSelfHosted,
		Metadata: app.InstallerMetadata{
			Description:      req.Metadata.Description,
			Name:             req.Name,
			CommunityURL:     req.Metadata.CommunityURL,
			HomepageURL:      req.Metadata.HomepageURL,
			DocumentationURL: req.Metadata.DocumentationURL,
			GithubURL:        req.Metadata.GithubURL,
			LogoURL:          req.Metadata.LogoURL,
			FaviconURL:       req.Metadata.FaviconURL,

			DemoURL:             generics.NewNullString(req.Metadata.DemoURL),
			OgImageURL:          generics.NewNullString(req.Metadata.OgImageURL),
			PostInstallMarkdown: generics.NewNullString(req.Metadata.PostInstallMarkdown),
			CopyrightMarkdown:   generics.NewNullString(req.Metadata.CopyrightMarkdown),
			FooterMarkdown:      generics.NewNullString(req.Metadata.FooterMarkdown),
		},
	}

	res := s.db.WithContext(ctx).Create(&installer)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create installer: %w", res.Error)
	}

	return &installer, nil
}
