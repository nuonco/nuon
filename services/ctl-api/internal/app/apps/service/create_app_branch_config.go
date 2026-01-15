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

type InstallGroupRequest struct {
	Name              string   `json:"name" validate:"required,min=1"`
	Order             int      `json:"order" validate:"min=0"`
	InstallIDs        []string `json:"install_ids"`
	RequiresApproval  bool     `json:"requires_approval"`
	RollbackOnFailure bool     `json:"rollback_on_failure"`
	MaxParallel       int      `json:"max_parallel" validate:"min=1"`
}

type CreateAppBranchConfigRequest struct {
	basicVCSConfigRequest
	InstallGroups []InstallGroupRequest `json:"install_groups"`
}

func (c *CreateAppBranchConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return err
	}

	if err := c.basicVCSConfigRequest.Validate(); err != nil {
		return err
	}

	// Validate install groups have unique orders
	orders := make(map[int]bool)
	for _, group := range c.InstallGroups {
		if orders[group.Order] {
			return stderr.ErrUser{
				Err:         fmt.Errorf("duplicate install group order: %d", group.Order),
				Description: "install groups must have unique order values",
			}
		}
		orders[group.Order] = true
	}

	return nil
}

// @ID						CreateAppBranchConfig
// @Summary				create an app branch config
// @Description.markdown	create_app_branch_config.md
// @Tags					apps
// @Accept					json
// @Param					req				body	CreateAppBranchConfigRequest	true	"Input"
// @Param					app_id			path	string							true	"app ID"
// @Param					app_branch_id	path	string							true	"app branch ID"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppBranchConfig
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/configs [post]
func (s *service) CreateAppBranchConfig(ctx *gin.Context) {
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

	appID := ctx.Param("app_id")
	appBranchID := ctx.Param("app_branch_id")

	var req CreateAppBranchConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// Verify branch exists and belongs to this org/app
	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Where(app.AppBranch{
			OrgID: org.ID,
			AppID: appID,
		}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
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

	// Convert request install groups to model
	installGroups := make([]app.AppBranchInstallGroup, len(req.InstallGroups))
	for i, g := range req.InstallGroups {
		maxParallel := g.MaxParallel
		if maxParallel == 0 {
			maxParallel = 5 // default
		}
		installGroups[i] = app.AppBranchInstallGroup{
			Name:              g.Name,
			Order:             g.Order,
			InstallIDs:        g.InstallIDs,
			RequiresApproval:  g.RequiresApproval,
			RollbackOnFailure: g.RollbackOnFailure,
			MaxParallel:       maxParallel,
		}
	}

	config, err := s.helpers.CreateAppBranchConfig(
		ctx,
		appBranchID,
		connectedGithubVCSConfig,
		publicGitVCSConfig,
		installGroups,
	)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app branch config: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, config)
}
