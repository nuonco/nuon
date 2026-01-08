package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateAppBranchRequest struct {
	basicVCSConfigRequest
	Name string `json:"name" validate:"required,min=1"`
}

func (c *CreateAppBranchRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return err
	}

	if err := c.basicVCSConfigRequest.Validate(); err != nil {
		return err
	}

	// App branches only support connected GitHub repos, not public repos (if VCS config is provided)
	if c.PublicGitVCSConfig != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("public git repos not supported for app branches"),
			Description: "App branches only support connected GitHub repositories. Please use a connected_github_vcs_config or omit VCS config entirely.",
		}
	}

	return nil
}

// @ID						CreateAppBranch
// @Description.markdown	create_app_branch.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppBranchRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppBranch
// @Router					/v1/apps/{app_id}/branches [post]
func (s *service) CreateAppBranch(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Feature flag checks
	if !org.Features[string(app.OrgFeatureAppBranches)] {
		ctx.Error(fmt.Errorf("app branches feature not enabled for this organization"))
		return
	}

	if !org.Features[string(app.OrgFeatureQueues)] {
		ctx.Error(fmt.Errorf("queues feature not enabled for this organization"))
		return
	}

	appID := ctx.Param("app_id")

	var req CreateAppBranchRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// Load app with org and VCS connections for lookup
	parentApp, err := s.getAppWithOrg(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	// Create VCS config using shared helpers
	connectedGithubVCSConfig, err := req.connectedGithubVCSConfig(ctx, parentApp, s.vcsHelpers)
	if err != nil {
		ctx.Error(fmt.Errorf("invalid connected github vcs config: %w", err))
		return
	}

	publicGitVCSConfig, err := req.publicGitVCSConfig(ctx, parentApp, s.vcsHelpers)
	if err != nil {
		ctx.Error(fmt.Errorf("invalid public git vcs config: %w", err))
		return
	}

	branch, err := s.helpers.CreateAppBranch(ctx, appID, req.Name, connectedGithubVCSConfig, publicGitVCSConfig)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app branch: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, branch)
}

func (s *service) getAppWithOrg(ctx *gin.Context, appID string) (*app.App, error) {
	var parentApp app.App
	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("Org.VCSConnections").
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}
	return &parentApp, nil
}
