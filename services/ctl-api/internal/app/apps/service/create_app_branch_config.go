package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type InstallGroupRequest struct {
	Name           string   `json:"name" validate:"required,min=1"`
	Order          int      `json:"order" validate:"min=0"`
	InstallIDs     []string `json:"install_ids"`
	UseForPreviews bool     `json:"use_for_previews"`

	// LabelSelector dynamically resolves installs at deploy time.
	// Mutually exclusive with InstallIDs.
	LabelSelector *labels.Selector `json:"label_selector,omitempty"`
}

type CreateAppBranchConfigRequest struct {
	vcshelpers.VCSConfigRequest

	InstallGroups []InstallGroupRequest `json:"install_groups"`
}

func (c *CreateAppBranchConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return err
	}

	if err := c.VCSConfigRequest.Validate(); err != nil {
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

		// InstallIDs and LabelSelector are mutually exclusive
		hasIDs := len(group.InstallIDs) > 0
		hasSelector := group.LabelSelector != nil && len(group.LabelSelector.MatchLabels) > 0
		if hasIDs && hasSelector {
			return stderr.ErrUser{
				Err:         fmt.Errorf("install group %q has both install_ids and label_selector", group.Name),
				Description: "install groups must use either install_ids or label_selector, not both",
			}
		}
		if !hasIDs && !hasSelector {
			return stderr.ErrUser{
				Err:         fmt.Errorf("install group %q has neither install_ids nor label_selector", group.Name),
				Description: "install groups must specify either install_ids or label_selector",
			}
		}
		if hasSelector {
			if err := group.LabelSelector.Validate(); err != nil {
				return stderr.ErrUser{
					Err:         fmt.Errorf("install group %q has invalid label_selector: %w", group.Name, err),
					Description: "label_selector must have non-empty match_labels",
				}
			}
		}
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
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppBranchConfig
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/configs [post]
func (s *service) CreateAppBranchConfig(ctx *gin.Context) {
	// Feature flag checks
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureAppBranches)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check feature: %w", err))
		return
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureAppBranches))
		return
	}

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
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

	// Validate that all app branches use the same repository BEFORE creating VCS configs
	branches, err := s.helpers.FetchAppBranchesWithConfigs(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to fetch app branches: %w", err))
		return
	}

	if err := s.helpers.ValidateSameRepo(branches, &req.VCSConfigRequest); err != nil {
		ctx.Error(err)
		return
	}

	// Load app with org and VCS connections for lookup
	parentApp, err := s.getAppWithOrg(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	// Build VCS configs (after validation passes)
	connectedGithubVCSConfig, err := s.vcsHelpers.BuildConnectedGithubVCSConfig(ctx, req.ConnectedGithubVCSConfig, parentApp.Org)
	if err != nil {
		ctx.Error(fmt.Errorf("invalid connected github vcs config: %w", err))
		return
	}

	publicGitVCSConfig, err := s.vcsHelpers.BuildPublicGitVCSConfig(ctx, req.PublicGitVCSConfig)
	if err != nil {
		ctx.Error(fmt.Errorf("invalid public git vcs config: %w", err))
		return
	}

	// Convert request install groups to model
	installGroups := make([]app.AppBranchInstallGroup, len(req.InstallGroups))
	for i, g := range req.InstallGroups {
		selector := g.LabelSelector
		if selector != nil && len(selector.MatchLabels) == 0 {
			selector = nil
		}
		installGroups[i] = app.AppBranchInstallGroup{
			Name:           g.Name,
			Order:          g.Order,
			InstallIDs:     g.InstallIDs,
			LabelSelector:  selector,
			UseForPreviews: g.UseForPreviews,
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
