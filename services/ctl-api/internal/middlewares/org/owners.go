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
	{
		prefix:  "/v1/runners/terraform-workspace",
		idParam: "workspace_id",
		resolve: ownedVia(func() any { return &app.TerraformWorkspace{} }),
	},
	{
		prefix:  "/v1/runners",
		idParam: "runner_id",
		resolve: resolveRunnerOwner,
	},
	{
		prefix:  "/v1/runner-jobs",
		idParam: "runner_job_id",
		resolve: ownedVia(func() any { return &app.RunnerJob{} }),
	},
	{
		prefix:  "/v1/log-streams",
		idParam: "log_stream_id",
		resolve: ownedVia(func() any { return &app.LogStream{} }),
	},
	{
		prefix:  "/v1/terraform-workspaces",
		idParam: "workspace_id",
		resolve: ownedVia(func() any { return &app.TerraformWorkspace{} }),
	},
	{
		prefix:  "/v1/queues",
		idParam: "queue_id",
		resolve: ownedVia(func() any { return &app.Queue{} }),
	},
	{
		prefix:  "/v1/policy-reports",
		idParam: "report_id",
		resolve: resolvePolicyReportOwner,
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

// ownerRefDepthLimit bounds the polymorphic owner walk; ownership graphs are
// shallow (log stream -> runner job -> deploy -> install is the deepest).
const ownerRefDepthLimit = 5

// ownedVia resolves a model with polymorphic OwnerID/OwnerType columns by
// walking owner references until an install, app, or org tier is reached.
func ownedVia(model func() any) func(*gin.Context, *gorm.DB, string, string) (string, string, error) {
	return func(ctx *gin.Context, db *gorm.DB, orgID, id string) (string, string, error) {
		var row struct {
			OwnerID   string
			OwnerType string
		}
		err := db.WithContext(ctx).
			Model(model()).
			Select("owner_id, owner_type").
			Where("org_id = ?", orgID).
			Where("id = ?", id).
			Take(&row).Error
		if err != nil {
			return "", "", err
		}
		return resolveOwnerRef(ctx, db, orgID, row.OwnerType, row.OwnerID, 0)
	}
}

func resolveWorkflowOwner(ctx *gin.Context, db *gorm.DB, orgID, id string) (string, string, error) {
	return ownedVia(func() any { return &app.Workflow{} })(ctx, db, orgID, id)
}

// resolveOwnerRef maps a polymorphic (owner_type, owner_id) reference to the
// owning install and/or app, walking intermediate owners (runner groups,
// workflows, deploys, ...) until a grantable tier is reached. Unknown owner
// types fail closed for grant-scoped accounts.
func resolveOwnerRef(ctx *gin.Context, db *gorm.DB, orgID, ownerType, ownerID string, depth int) (string, string, error) {
	if depth > ownerRefDepthLimit {
		return "", "", fmt.Errorf("owner reference chain exceeds depth limit at %s/%s", ownerType, ownerID)
	}

	selectOwner := func(model any) (string, string, error) {
		var row struct {
			OwnerID   string
			OwnerType string
		}
		err := db.WithContext(ctx).
			Model(model).
			Select("owner_id, owner_type").
			Where("org_id = ?", orgID).
			Where("id = ?", ownerID).
			Take(&row).Error
		if err != nil {
			return "", "", err
		}
		return resolveOwnerRef(ctx, db, orgID, row.OwnerType, row.OwnerID, depth+1)
	}

	switch ownerType {
	case "orgs", "org":
		if ownerID != orgID {
			return "", "", fmt.Errorf("owner org %s does not match request org", ownerID)
		}
		return "", "", nil
	case "installs", "install":
		var row struct{ AppID string }
		err := db.WithContext(ctx).
			Model(&app.Install{}).
			Select("app_id").
			Where("org_id = ?", orgID).
			Where("id = ?", ownerID).
			Take(&row).Error
		if err != nil {
			return "", "", err
		}
		return ownerID, row.AppID, nil
	case "apps", "app":
		return "", ownerID, nil
	case "app_branches":
		var row struct{ AppID string }
		err := db.WithContext(ctx).
			Model(&app.AppBranch{}).
			Select("app_id").
			Where("org_id = ?", orgID).
			Where("id = ?", ownerID).
			Take(&row).Error
		if err != nil {
			return "", "", err
		}
		return "", row.AppID, nil
	case "components":
		return appOwned(func() any { return &app.Component{} })(ctx, db, orgID, ownerID)
	case "component_builds":
		return resolveComponentBuildOwner(ctx, db, orgID, ownerID)
	case "install_deploys":
		var row struct{ InstallID string }
		err := db.WithContext(ctx).
			Model(&app.InstallDeploy{}).
			Select("install_components.install_id").
			Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id AND install_components.deleted_at = 0").
			Where("install_deploys.org_id = ?", orgID).
			Where("install_deploys.id = ?", ownerID).
			Take(&row).Error
		if err != nil {
			return "", "", err
		}
		return resolveOwnerRef(ctx, db, orgID, "installs", row.InstallID, depth+1)
	case "install_sandbox_runs":
		return installOwned(func() any { return &app.InstallSandboxRun{} }, "install_sandbox_runs")(ctx, db, orgID, ownerID)
	case "install_workflows":
		return selectOwner(&app.Workflow{})
	case "install_workflow_steps":
		var row struct{ InstallWorkflowID string }
		err := db.WithContext(ctx).
			Model(&app.WorkflowStep{}).
			Select("install_workflow_id").
			Where("org_id = ?", orgID).
			Where("id = ?", ownerID).
			Take(&row).Error
		if err != nil {
			return "", "", err
		}
		return selectOwnerByID(ctx, db, orgID, &app.Workflow{}, row.InstallWorkflowID, depth+1)
	case "runners":
		var row struct{ RunnerGroupID string }
		err := db.WithContext(ctx).
			Model(&app.Runner{}).
			Select("runner_group_id").
			Where("org_id = ?", orgID).
			Where("id = ?", ownerID).
			Take(&row).Error
		if err != nil {
			return "", "", err
		}
		return selectOwnerByID(ctx, db, orgID, &app.RunnerGroup{}, row.RunnerGroupID, depth+1)
	case "runner_groups":
		return selectOwner(&app.RunnerGroup{})
	case "runner_jobs":
		return selectOwner(&app.RunnerJob{})
	default:
		return "", "", fmt.Errorf("unresolvable owner type %q", ownerType)
	}
}

// selectOwnerByID fetches a model's polymorphic owner columns by an explicit
// ID (rather than the in-scope ownerID) and continues the walk.
func selectOwnerByID(ctx *gin.Context, db *gorm.DB, orgID string, model any, id string, depth int) (string, string, error) {
	var row struct {
		OwnerID   string
		OwnerType string
	}
	err := db.WithContext(ctx).
		Model(model).
		Select("owner_id, owner_type").
		Where("org_id = ?", orgID).
		Where("id = ?", id).
		Take(&row).Error
	if err != nil {
		return "", "", err
	}
	return resolveOwnerRef(ctx, db, orgID, row.OwnerType, row.OwnerID, depth)
}

// resolveRunnerOwner authorizes a runner via its group's owner: install
// runners inherit their install's chain, org runners resolve to the bare org
// tier (org-wide permission required).
func resolveRunnerOwner(ctx *gin.Context, db *gorm.DB, orgID, id string) (string, string, error) {
	var row struct{ RunnerGroupID string }
	err := db.WithContext(ctx).
		Model(&app.Runner{}).
		Select("runner_group_id").
		Where("org_id = ?", orgID).
		Where("id = ?", id).
		Take(&row).Error
	if err != nil {
		return "", "", err
	}
	return selectOwnerByID(ctx, db, orgID, &app.RunnerGroup{}, row.RunnerGroupID, 0)
}

func resolvePolicyReportOwner(ctx *gin.Context, db *gorm.DB, orgID, id string) (string, string, error) {
	var row struct {
		AppID     string
		InstallID *string
	}
	err := db.WithContext(ctx).
		Model(&app.PolicyReport{}).
		Select("app_id, install_id").
		Where("org_id = ?", orgID).
		Where("id = ?", id).
		Take(&row).Error
	if err != nil {
		return "", "", err
	}
	installID := ""
	if row.InstallID != nil {
		installID = *row.InstallID
	}
	return installID, row.AppID, nil
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
