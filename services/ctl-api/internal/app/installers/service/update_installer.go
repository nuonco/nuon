package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type UpdateInstallerRequest struct {
	AppIDs []string `validate:"required" json:"app_ids"`
	Name   string   `validate:"required" json:"name"`

	Metadata struct {
		Description      string `validate:"required" json:"description"`
		DocumentationURL string `validate:"required" json:"documentation_url"`
		LogoURL          string `validate:"required" json:"logo_url"`
		GithubURL        string `validate:"required" json:"github_url"`
		HomepageURL      string `validate:"required" json:"homepage_url"`
		CommunityURL     string `validate:"required" json:"community_url"`
		FaviconURL       string `validate:"required" json:"favicon_url"`

		DemoURL             string `json:"demo_url"`
		OgImageURL          string `json:"og_image_url" `
		PostInstallMarkdown string `json:"post_install_markdown"`
		FooterMarkdown      string `json:"footer_markdown"`
		CopyrightMarkdown   string `json:"copyright_markdown"`
	} `json:"metadata"`
}

func (c *UpdateInstallerRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

//	@ID						UpdateInstaller
//	@Summary				update an installer
//	@Description.markdown	update_installer.md
//	@Tags					installers
//	@Accept					json
//	@Param					req	body	UpdateInstallerRequest	true	"Input"
//	@Produce				json
//	@Param					installer_id	path	string	true	"installer ID"
//	@Security				APIKey
//	@Security				OrgID
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				201	{object}	app.Installer
//	@Router					/v1/installers/{installer_id} [PATCH]
func (s *service) UpdateInstaller(ctx *gin.Context) {
	installerID := ctx.Param("installer_id")

	var req UpdateInstallerRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	installer, err := s.updateInstaller(ctx, installerID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update app installer: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, installer)
}

func (s *service) updateInstaller(ctx context.Context, installerID string, req *UpdateInstallerRequest) (*app.Installer, error) {
	installer, err := s.getInstaller(ctx, installerID)
	if err != nil {
		return nil, err
	}

	updates := app.InstallerMetadata{
		Description:         req.Metadata.Description,
		Name:                req.Name,
		CommunityURL:        req.Metadata.CommunityURL,
		HomepageURL:         req.Metadata.HomepageURL,
		DocumentationURL:    req.Metadata.DocumentationURL,
		GithubURL:           req.Metadata.GithubURL,
		LogoURL:             req.Metadata.LogoURL,
		FaviconURL:          req.Metadata.FaviconURL,
		DemoURL:             generics.NewNullString(req.Metadata.DemoURL),
		OgImageURL:          generics.NewNullString(req.Metadata.OgImageURL),
		PostInstallMarkdown: generics.NewNullString(req.Metadata.PostInstallMarkdown),
		CopyrightMarkdown:   generics.NewNullString(req.Metadata.CopyrightMarkdown),
		FooterMarkdown:      generics.NewNullString(req.Metadata.FooterMarkdown),
	}

	metadata := app.InstallerMetadata{}
	res := s.db.WithContext(ctx).
		Model(&metadata).
		Where("installer_id = ?", installerID).
		Updates(updates)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update app installer: %w", res.Error)
	}

	// update apps
	var apps []app.App
	for _, appID := range req.AppIDs {
		apps = append(apps, app.App{
			ID: appID,
		})
	}
	err = s.db.WithContext(ctx).Model(&app.Installer{
		ID: installerID,
	}).Association("Apps").Replace(apps)
	if err != nil {
		return nil, fmt.Errorf("unable to find installer: %w", err)
	}

	return installer, nil
}
