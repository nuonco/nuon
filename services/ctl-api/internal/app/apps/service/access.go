package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	installaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	orgaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const authorizedAppKey = "nuon.apps.authorized_app"

func (s *service) appResources(ctx *gin.Context) (*authz.Resources, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return access.Apps.In(s.db, orgID)
}

func (s *service) appTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	a, err := s.findApp(ctx, orgID, ctx.Param("app_id"))
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := s.appResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	target, err := resources.Resolve(ctx, a.ID)
	if err != nil {
		return permissions.Target{}, err
	}
	ctx.Set(authorizedAppKey, a)
	return target, nil
}

func (s *service) exactAppTarget(ctx *gin.Context) (permissions.Target, error) {
	resources, err := s.appResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, ctx.Param("app_id"))
}

func (s *service) appOrgTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := orgaccess.Orgs.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, orgID)
}

func (s *service) appRule(operation permissions.PermissionID) require.Rule {
	return require.Point(operation, s.exactAppTarget)
}

func (s *service) namedAppRule(operation permissions.PermissionID) require.Rule {
	return require.Point(operation, s.appTarget)
}

func (s *service) branchResources(ctx *gin.Context) (*authz.Resources, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return access.Branches.In(s.db, orgID)
}

func (s *service) installGroupRunTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := access.InstallGroupRuns.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, ctx.Param("install_group_run_id"))
}

func (s *service) childConfigTarget(ctx *gin.Context) (permissions.Target, error) {
	target, err := s.exactAppTarget(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var req struct {
		AppConfigID string `json:"app_config_id"`
	}
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return permissions.Target{}, stderr.NewInvalidRequest(err)
	}
	if req.AppConfigID != "" {
		if _, err := s.getAppConfig(ctx, target.OrgID, target.ID, req.AppConfigID); err != nil {
			return permissions.Target{}, err
		}
	}
	return target, nil
}

func (s *service) childConfigRule() require.Rule {
	return require.Point(permissions.AppConfigWrite, s.childConfigTarget)
}

func (s *service) configApplyRequirements(ctx *gin.Context) ([]authz.Requirement, error) {
	target, err := s.childConfigTarget(ctx)
	if err != nil {
		return nil, err
	}
	var installs []app.Install
	if err := s.db.WithContext(ctx).Where(app.Install{AppID: target.ID, OrgID: target.OrgID}).Find(&installs).Error; err != nil {
		return nil, err
	}
	requirements := installConfigRequirements(ctx, target, installs)
	return append(requirements, authz.Requirement{Operation: permissions.AppConfigWrite, Target: target}), nil
}

func (s *service) branchDeleteRequirements(ctx *gin.Context) ([]authz.Requirement, error) {
	target, err := s.exactAppTarget(ctx)
	if err != nil {
		return nil, err
	}
	var branch app.AppBranch
	if err := s.db.WithContext(ctx).Where(app.AppBranch{ID: ctx.Param("app_branch_id"), AppID: target.ID, OrgID: target.OrgID}).First(&branch).Error; err != nil {
		return nil, err
	}
	var connections []app.InstallAppBranchConnection
	if err := s.db.WithContext(ctx).Where(app.InstallAppBranchConnection{AppBranchID: branch.ID, OrgID: target.OrgID, Active: true}).Find(&connections).Error; err != nil {
		return nil, err
	}
	resources, err := installaccess.Resources(s.db, target.OrgID)
	if err != nil {
		return nil, err
	}
	requirements := []authz.Requirement{{Operation: permissions.AppBranchDelete, Target: target}}
	ids := make([]string, 0, len(connections))
	for _, connection := range connections {
		install, err := resources.Resolve(ctx, connection.InstallID)
		if err != nil {
			return nil, err
		}
		if install.AppID != target.ID {
			return nil, gorm.ErrRecordNotFound
		}
		requirements = append(requirements, authz.Requirement{Operation: permissions.InstallConfigUpdate, Target: install})
		ids = append(ids, connection.ID)
	}
	ctx.Set("nuon.apps.branch_connection_ids", ids)
	return requirements, nil
}

func (s *service) branchConfigSyncRequirements(ctx *gin.Context) ([]authz.Requirement, error) {
	target, err := s.exactAppTarget(ctx)
	if err != nil {
		return nil, err
	}
	var branch app.AppBranch
	if err := s.db.WithContext(ctx).Where(app.AppBranch{ID: ctx.Param("app_branch_id"), AppID: target.ID, OrgID: target.OrgID}).First(&branch).Error; err != nil {
		return nil, err
	}
	var req TriggerInstallConfigSyncRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return nil, stderr.NewInvalidRequest(err)
	}
	var installs []app.Install
	query := s.db.WithContext(ctx).Where(app.Install{AppID: target.ID, OrgID: target.OrgID})
	if req.InstallName != "" {
		query = query.Where(app.Install{Name: req.InstallName})
	}
	if err := query.Find(&installs).Error; err != nil {
		return nil, err
	}
	return installConfigRequirements(ctx, target, installs), nil
}

func (s *service) configInstallRequirements(ctx *gin.Context) ([]authz.Requirement, error) {
	target, err := s.exactAppTarget(ctx)
	if err != nil {
		return nil, err
	}
	configID := ctx.Param("config_id")
	if configID == "" {
		configID = ctx.Param("app_config_id")
	}
	if _, err := s.getAppConfig(ctx, target.OrgID, target.ID, configID); err != nil {
		return nil, err
	}
	var req UpdateAppConfigInstallsRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return nil, stderr.NewInvalidRequest(err)
	}
	var installs []app.Install
	query := s.db.WithContext(ctx).Where(app.Install{AppID: target.ID, OrgID: target.OrgID})
	if !req.UpdateAll {
		query = query.Where("id IN ?", req.InstallIDs)
	}
	if err := query.Find(&installs).Error; err != nil {
		return nil, err
	}
	return installConfigRequirements(ctx, target, installs), nil
}

func installConfigRequirements(ctx *gin.Context, target permissions.Target, installs []app.Install) []authz.Requirement {
	requirements := []authz.Requirement{{Operation: permissions.AppConfigApply, Target: target}}
	ids := make([]string, 0, len(installs))
	for _, install := range installs {
		ids = append(ids, install.ID)
		requirements = append(requirements, authz.Requirement{Operation: permissions.InstallConfigUpdate, Target: permissions.Target{ResolvedTarget: permissions.ResolvedTarget{
			Scope: permissions.Scope{Kind: permissions.KindInstall, OrgID: target.OrgID, ID: install.ID}, AppID: target.ID, InstallID: install.ID,
		}}})
	}
	ctx.Set("nuon.apps.config_install_ids", ids)
	return requirements
}
