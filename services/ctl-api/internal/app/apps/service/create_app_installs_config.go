package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateAppInstallsConfigRequest struct {
	VCSType         string  `json:"vcs_type" validate:"required,oneof=connected public"`
	VCSConnectionID *string `json:"vcs_connection_id,omitempty"`
	Repo            string  `json:"repo" validate:"required"`
	Branch          string  `json:"branch" validate:"required"`
	Directory       string  `json:"directory,omitempty"`
}

// @ID						CreateAppInstallsConfig
// @Summary				create a new installs config for an app
// @Description			Creates a new installs config record (source=ui). The latest record is always used.
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					app_id	path	string							true	"app ID"
// @Param					req		body	CreateAppInstallsConfigRequest	true	"Input"
// @Success				201	{object}	app.AppInstallsConfig
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/apps/{app_id}/installs-configs [post]
func (s *service) CreateAppInstallsConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !s.requireInstallSyncing(ctx) {
		return
	}

	appID := ctx.Param("app_id")

	var req CreateAppInstallsConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	if err := s.v.Struct(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	var a app.App
	if err := s.db.WithContext(ctx).
		Preload("Org").
		Preload("Org.VCSConnections").
		Where(app.App{ID: appID, OrgID: org.ID}).
		First(&a).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	dir := req.Directory
	if dir == "" {
		dir = "."
	}

	cfg := app.AppInstallsConfig{
		AppID:           appID,
		VCSType:         req.VCSType,
		VCSConnectionID: req.VCSConnectionID,
		Repo:            req.Repo,
		Branch:          req.Branch,
		Directory:       dir,
		Source:          "ui",
	}

	if err := s.db.WithContext(ctx).Create(&cfg).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to create installs config: %w", err))
		return
	}

	switch req.VCSType {
	case "connected":
		connectedCfg, err := s.vcsHelpers.BuildConnectedGithubVCSConfig(ctx, &vcshelpers.ConnectedGithubVCSConfigRequest{
			Repo:      req.Repo,
			Branch:    req.Branch,
			Directory: dir,
		}, a.Org)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to build connected VCS config: %w", err))
			return
		}
		if err := s.vcsHelpers.AttachVCSConfigs(ctx, vcshelpers.AttachVCSConfigsParams{
			OwnerID:            cfg.ID,
			OwnerType:          &cfg,
			ConnectedGithubVCS: connectedCfg,
		}); err != nil {
			ctx.Error(fmt.Errorf("unable to attach connected VCS config: %w", err))
			return
		}
		cfg.ConnectedGithubVCSConfig = connectedCfg
	case "public":
		publicCfg, err := s.vcsHelpers.BuildPublicGitVCSConfig(ctx, &vcshelpers.PublicGitVCSConfigRequest{
			Repo:      req.Repo,
			Branch:    req.Branch,
			Directory: dir,
		})
		if err != nil {
			ctx.Error(fmt.Errorf("unable to build public VCS config: %w", err))
			return
		}
		if err := s.vcsHelpers.AttachVCSConfigs(ctx, vcshelpers.AttachVCSConfigsParams{
			OwnerID:      cfg.ID,
			OwnerType:    &cfg,
			PublicGitVCS: publicCfg,
		}); err != nil {
			ctx.Error(fmt.Errorf("unable to attach public VCS config: %w", err))
			return
		}
		cfg.PublicGitVCSConfig = publicCfg
	}

	ctx.JSON(http.StatusCreated, cfg)
}
