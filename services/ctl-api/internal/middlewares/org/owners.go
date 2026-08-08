package org

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

// ownedResourceRoute maps a route prefix to a child resource (build, workflow,
// component, ...) whose owning install or app is resolved from the database so
// the request authorizes exactly like a request against the owner itself.
// Children never become grantable types — they inherit the owner's chain.
//
// Prefix-keyed like orgResourceRoutes so ambiguous param names can't collide;
// entries are matched in order, so more specific prefixes come first.
type ownedResourceRoute struct {
	prefix  string
	idParam string
	// resolve returns the owning install and/or app for the child, scoped to
	// the org (owner columns only — not a full fetch). installID is empty for
	// app-owned children; appID is set whenever known so the chain includes
	// every tier.
	resolve func(ctx *gin.Context, db *gorm.DB, orgID, id string) (installID, appID string, err error)
}

var ownedResourceRoutes = []ownedResourceRoute{
	{
		prefix:  "/v1/action-workflows/configs",
		idParam: "action_workflow_config_id",
		resolve: appOwned(func() any { return &app.ActionWorkflowConfig{} }),
	},
	{
		prefix:  "/v1/action-workflows",
		idParam: "action_workflow_id",
		resolve: appOwned(func() any { return &app.ActionWorkflow{} }),
	},
	{
		prefix:  "/v1/components/builds",
		idParam: "build_id",
		resolve: resolveComponentBuildOwner,
	},
	{
		prefix:  "/v1/components",
		idParam: "component_id",
		resolve: appOwned(func() any { return &app.Component{} }),
	},
	{
		prefix:  "/v1/install-workflows",
		idParam: "install_workflow_id",
		resolve: resolveWorkflowOwner,
	},
	{
		prefix:  "/v1/workflows",
		idParam: "workflow_id",
		resolve: resolveWorkflowOwner,
	},
	{
		prefix:  "/v1/installs/sandbox-runs",
		idParam: "run_id",
		resolve: installOwned(func() any { return &app.InstallSandboxRun{} }, "install_sandbox_runs"),
	},
	{
		prefix:  "/v1/installs/stacks",
		idParam: "stack_id",
		resolve: installOwned(func() any { return &app.InstallStack{} }, "install_stacks"),
	},
}

func matchOwnedResourceRoute(fullPath string) (ownedResourceRoute, bool) {
	for _, r := range ownedResourceRoutes {
		if strings.HasPrefix(fullPath, r.prefix) {
			return r, true
		}
	}
	return ownedResourceRoute{}, false
}

// typeOnlyCreateRoutes maps parentless create routes to the tier being
// created; they resolve to a type-only link satisfiable by that tier's
// wildcard grant or anything above it.
var typeOnlyCreateRoutes = map[string]app.GrantResourceType{
	"POST /v1/installs": app.GrantResourceTypeInstall,
	"POST /v1/apps":     app.GrantResourceTypeApp,
}

func matchTypeOnlyCreateRoute(method, fullPath string) (app.GrantResourceType, bool) {
	t, ok := typeOnlyCreateRoutes[routeKey(method, fullPath)]
	return t, ok
}

// ownerChain builds the walk-up chain for a resolved owner: [install, app,
// org] when the child is install-owned, [app, org] when app-owned.
func ownerChain(orgID, installID, appID string) []authz.Link {
	chain := make([]authz.Link, 0, 3)
	if installID != "" {
		chain = append(chain, authz.Link{Type: string(app.GrantResourceTypeInstall), ID: installID})
	}
	if appID != "" {
		chain = append(chain, authz.Link{Type: string(app.GrantResourceTypeApp), ID: appID})
	}
	return append(chain, authz.Link{Type: string(app.GrantResourceTypeOrg), ID: orgID})
}

func appOwned(model func() any) func(*gin.Context, *gorm.DB, string, string) (string, string, error) {
	return func(ctx *gin.Context, db *gorm.DB, orgID, id string) (string, string, error) {
		var row struct{ AppID string }
		err := db.WithContext(ctx).
			Model(model()).
			Select("app_id").
			Where("org_id = ?", orgID).
			Where("id = ?", id).
			Take(&row).Error
		if err != nil {
			return "", "", err
		}
		return "", row.AppID, nil
	}
}

func installOwned(model func() any, table string) func(*gin.Context, *gorm.DB, string, string) (string, string, error) {
	return func(ctx *gin.Context, db *gorm.DB, orgID, id string) (string, string, error) {
		var row struct {
			InstallID string
			AppID     string
		}
		err := db.WithContext(ctx).
			Model(model()).
			Select(table+".install_id, installs.app_id").
			Joins("JOIN installs ON installs.id = "+table+".install_id AND installs.deleted_at = 0").
			Where(table+".org_id = ?", orgID).
			Where(table+".id = ?", id).
			Take(&row).Error
		if err != nil {
			return "", "", err
		}
		return row.InstallID, row.AppID, nil
	}
}

// resolveWorkflowOwner maps a workflow to its polymorphic owner. Workflows
// owned by anything other than an install or app (e.g. app branches) fail
// closed for grant-scoped accounts until their tier is resolvable.
func resolveWorkflowOwner(ctx *gin.Context, db *gorm.DB, orgID, id string) (string, string, error) {
	var wf struct {
		OwnerID   string
		OwnerType string
	}
	err := db.WithContext(ctx).
		Model(&app.Workflow{}).
		Select("owner_id, owner_type").
		Where("org_id = ?", orgID).
		Where("id = ?", id).
		Take(&wf).Error
	if err != nil {
		return "", "", err
	}

	switch wf.OwnerType {
	case "installs":
		var row struct{ AppID string }
		err := db.WithContext(ctx).
			Model(&app.Install{}).
			Select("app_id").
			Where("org_id = ?", orgID).
			Where("id = ?", wf.OwnerID).
			Take(&row).Error
		if err != nil {
			return "", "", err
		}
		return wf.OwnerID, row.AppID, nil
	case "apps":
		return "", wf.OwnerID, nil
	default:
		return "", "", fmt.Errorf("workflow %s has unresolvable owner type %q", id, wf.OwnerType)
	}
}

func resolveComponentBuildOwner(ctx *gin.Context, db *gorm.DB, orgID, id string) (string, string, error) {
	var row struct{ AppID string }
	err := db.WithContext(ctx).
		Model(&app.ComponentBuild{}).
		Select("components.app_id").
		Joins("JOIN component_config_connections ON component_config_connections.id = component_builds.component_config_connection_id AND component_config_connections.deleted_at = 0").
		Joins("JOIN components ON components.id = component_config_connections.component_id AND components.deleted_at = 0").
		Where("component_builds.org_id = ?", orgID).
		Where("component_builds.id = ?", id).
		Take(&row).Error
	if err != nil {
		return "", "", err
	}
	return "", row.AppID, nil
}
