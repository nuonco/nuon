package authz

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (c *Client) ResolveTarget(ctx context.Context, scope permissions.Scope) (permissions.ResolvedTarget, error) {
	return resolveTarget(c.db.WithContext(ctx), scope)
}

func resolveTarget(db *gorm.DB, scope permissions.Scope) (permissions.ResolvedTarget, error) {
	if err := scope.ValidateShape(); err != nil {
		return permissions.ResolvedTarget{}, err
	}
	var org struct{ ID string }
	if err := db.Model(&app.Org{}).Select("id").Where(app.Org{ID: scope.OrgID}).Take(&org).Error; err != nil {
		return permissions.ResolvedTarget{}, fmt.Errorf("resolve target tenant: %w", err)
	}
	target, err := resolveTenantTarget(db, scope)
	if err != nil {
		return permissions.ResolvedTarget{}, err
	}
	if err := target.ValidateShape(); err != nil {
		return permissions.ResolvedTarget{}, err
	}
	return target, nil
}

func resolveTenantTarget(db *gorm.DB, scope permissions.Scope) (permissions.ResolvedTarget, error) {
	target := permissions.ResolvedTarget{Scope: scope}
	switch scope.Kind {
	case permissions.KindOrg:
		return target, nil
	case permissions.KindApp:
		var row struct{ ID string }
		if err := db.Model(&app.App{}).Select("id").Where(app.App{ID: scope.ID, OrgID: scope.OrgID}).Take(&row).Error; err != nil {
			return permissions.ResolvedTarget{}, fmt.Errorf("resolve app owner: %w", err)
		}
		target.AppID = row.ID
		return target, nil
	case permissions.KindInstall, permissions.KindStack, permissions.KindTelemetry:
		var row struct{ AppID string }
		if err := db.Model(&app.Install{}).Scopes(scopes.WithDisableViews).Select("app_id").
			Where(app.Install{ID: scope.ID, OrgID: scope.OrgID}).Take(&row).Error; err != nil {
			return permissions.ResolvedTarget{}, fmt.Errorf("resolve install owner: %w", err)
		}
		parent, err := resolveTarget(db, permissions.Scope{Kind: permissions.KindApp, OrgID: scope.OrgID, ID: row.AppID})
		if err != nil {
			return permissions.ResolvedTarget{}, err
		}
		target.AppID, target.InstallID = parent.AppID, scope.ID
		return target, nil
	case permissions.KindRunner:
		var runner struct{ RunnerGroupID string }
		if err := db.Model(&app.Runner{}).Select("runner_group_id").
			Where(app.Runner{ID: scope.ID, OrgID: scope.OrgID}).Take(&runner).Error; err != nil {
			return permissions.ResolvedTarget{}, fmt.Errorf("resolve runner group: %w", err)
		}
		if runner.RunnerGroupID == "" {
			return permissions.ResolvedTarget{}, fmt.Errorf("runner requires a group")
		}
		var group struct {
			OwnerID, OwnerType string
			Type               app.RunnerGroupType
		}
		if err := db.Model(&app.RunnerGroup{}).Select("owner_id", "owner_type", "type").
			Where(app.RunnerGroup{ID: runner.RunnerGroupID, OrgID: scope.OrgID}).Take(&group).Error; err != nil {
			return permissions.ResolvedTarget{}, fmt.Errorf("resolve runner group owner: %w", err)
		}
		if group.Type != app.RunnerGroupTypeInstall || group.OwnerType != plugins.TableName(db, app.Install{}) {
			return permissions.ResolvedTarget{}, fmt.Errorf("runner requires an install-owned group")
		}
		parent, err := resolveTarget(db, permissions.Scope{Kind: permissions.KindInstall, OrgID: scope.OrgID, ID: group.OwnerID})
		if err != nil {
			return permissions.ResolvedTarget{}, err
		}
		target.AppID, target.InstallID = parent.AppID, parent.InstallID
		return target, nil
	case permissions.KindRunnerJob, permissions.KindLogStream:
		var parent permissions.ResolvedTarget
		var err error
		if scope.Kind == permissions.KindRunnerJob {
			parent, err = resolveJobOwner(db, scope.OrgID, scope.ID)
		} else {
			parent, err = resolveLogStreamOwner(db, scope.OrgID, scope.ID)
		}
		if err != nil {
			return permissions.ResolvedTarget{}, err
		}
		target.AppID, target.InstallID = parent.AppID, parent.InstallID
		return target, nil
	case permissions.KindPhoneHome:
		var version struct{ InstallID, InstallStackID string }
		if err := db.Model(&app.InstallStackVersion{}).Select("install_id", "install_stack_id").
			Where(app.InstallStackVersion{ID: scope.ID, OrgID: scope.OrgID}).Take(&version).Error; err != nil {
			return permissions.ResolvedTarget{}, fmt.Errorf("resolve phone-home version: %w", err)
		}
		parent, err := resolveStackOwner(db, scope.OrgID, version.InstallStackID)
		if err != nil {
			return permissions.ResolvedTarget{}, err
		}
		if version.InstallID != parent.InstallID {
			return permissions.ResolvedTarget{}, fmt.Errorf("stack version and stack have different install owners")
		}
		target.AppID, target.InstallID = parent.AppID, parent.InstallID
		return target, nil
	default:
		return permissions.ResolvedTarget{}, fmt.Errorf("ownership resolver does not support resource kind %q", scope.Kind)
	}
}

func resolveStackOwner(db *gorm.DB, orgID, stackID string) (permissions.ResolvedTarget, error) {
	if err := (permissions.Scope{Kind: permissions.KindStack, OrgID: orgID, ID: stackID}).ValidateShape(); err != nil {
		return permissions.ResolvedTarget{}, err
	}
	var stack struct{ InstallID string }
	if err := db.Model(&app.InstallStack{}).Select("install_id").
		Where(app.InstallStack{ID: stackID, OrgID: orgID}).Take(&stack).Error; err != nil {
		return permissions.ResolvedTarget{}, fmt.Errorf("resolve stack owner: %w", err)
	}
	return resolveTarget(db, permissions.Scope{Kind: permissions.KindInstall, OrgID: orgID, ID: stack.InstallID})
}
