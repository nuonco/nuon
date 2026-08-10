package org

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
// Resolution is always a single org-scoped point read: tables either carry a
// direct owner column (components, policy_reports) or denormalized
// install_id/app_id ancestry columns populated at creation by
// app.ResolveOwnerAncestry and backfilled by migration 124. There is no
// recursive owner walk.
//
// Prefix-keyed like orgResourceRoutes so ambiguous param names can't collide;
// entries are matched in order, so more specific prefixes come first.
type ownedResourceRoute struct {
	prefix  string
	idParam string
	// resolve returns the owning install and/or app for the child, scoped to
	// the org (owner columns only — not a full fetch). installID is empty for
	// app-owned children; appID is set whenever known so the chain includes
	// every tier. Both empty means the child is an org-tier resource.
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
		resolve: denormOwned("install_workflows"),
	},
	{
		prefix:  "/v1/workflows",
		idParam: "workflow_id",
		resolve: denormOwned("install_workflows"),
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
		resolve: denormOwned("terraform_workspaces"),
	},
	{
		prefix:  "/v1/runners",
		idParam: "runner_id",
		resolve: resolveRunnerOwner,
	},
	{
		prefix:  "/v1/runner-jobs",
		idParam: "runner_job_id",
		resolve: denormOwned("runner_jobs"),
	},
	{
		prefix:  "/v1/log-streams",
		idParam: "log_stream_id",
		resolve: denormOwned("log_streams"),
	},
	{
		prefix:  "/v1/terraform-workspaces",
		idParam: "workspace_id",
		resolve: denormOwned("terraform_workspaces"),
	},
	{
		prefix:  "/v1/queues",
		idParam: "queue_id",
		resolve: denormOwned("queues"),
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

// queryOwnedRoute resolves a child named by a query parameter rather than a
// path param: the terraform HTTP backend protocol addresses its workspace as
// ?workspace_id=... on a fixed URL.
type queryOwnedRoute struct {
	queryParam string
	resolve    func(ctx *gin.Context, db *gorm.DB, orgID, id string) (installID, appID string, err error)
}

var queryOwnedRoutes = map[string]queryOwnedRoute{
	"GET /v1/terraform-backend":  {queryParam: "workspace_id", resolve: denormOwned("terraform_workspaces")},
	"POST /v1/terraform-backend": {queryParam: "workspace_id", resolve: denormOwned("terraform_workspaces")},
}

func matchQueryOwnedRoute(method, fullPath string) (queryOwnedRoute, bool) {
	r, ok := queryOwnedRoutes[routeKey(method, fullPath)]
	return r, ok
}

// bodyOwnedCreateRoutes name the created resource's owner in the JSON request
// body ({owner_type, owner_id}) — the terraform workspace creates, whose
// owners are install-tier (install_components, install_sandboxes). The chain
// is the body owner's ancestry, so an install grantee can create workspaces
// for their install; unknown owners resolve to the org tier (org-wide
// permission required).
var bodyOwnedCreateRoutes = map[string]struct{}{
	"POST /v1/terraform-workspace":  {},
	"POST /v1/terraform-workspaces": {},
}

func isBodyOwnedCreateRoute(method, fullPath string) bool {
	_, ok := bodyOwnedCreateRoutes[routeKey(method, fullPath)]
	return ok
}

// resolveBodyOwnerChain reads the request body (restoring it for the
// handler's own bind), extracts the polymorphic owner reference, and builds
// the owner's chain. Only ever invoked on the deferred-grantee path for
// routes in bodyOwnedCreateRoutes — org-wide callers never touch the body.
func resolveBodyOwnerChain(ctx *gin.Context, db *gorm.DB, orgID string) ([]authz.Link, error) {
	raw, err := ctx.GetRawData()
	if err != nil {
		return nil, err
	}
	ctx.Request.Body = io.NopCloser(bytes.NewReader(raw))

	var ref struct {
		OwnerID   string `json:"owner_id"`
		OwnerType string `json:"owner_type"`
	}
	if err := json.Unmarshal(raw, &ref); err != nil {
		return nil, fmt.Errorf("unable to parse owner from request body: %w", err)
	}

	installID, appID, err := app.ResolveOwnerAncestry(db.WithContext(ctx), orgID, ref.OwnerType, ref.OwnerID)
	if err != nil {
		return nil, err
	}
	var install, appTier string
	if installID != nil {
		install = *installID
	}
	if appID != nil {
		appTier = *appID
	}
	return ownerChain(orgID, install, appTier), nil
}

// ownerChain builds the walk-up chain for a resolved owner: [install, app,
// org] when the child is install-owned, [app, org] when app-owned, and the
// bare [org] for org-tier children (only org-wide permissions satisfy it).
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

// denormOwned resolves a child through its denormalized install_id/app_id
// ancestry columns. Both NULL means an org-tier resource (management-runner
// jobs, org queues, unresolved legacy rows) — the chain collapses to [org],
// failing closed for grant-scoped accounts.
func denormOwned(table string) func(*gin.Context, *gorm.DB, string, string) (string, string, error) {
	return func(ctx *gin.Context, db *gorm.DB, orgID, id string) (string, string, error) {
		var row struct {
			InstallID *string
			AppID     *string
		}
		err := db.WithContext(ctx).Table(table).
			Select("install_id, app_id").
			Where("deleted_at = 0").
			Where("org_id = ?", orgID).
			Where("id = ?", id).
			Take(&row).Error
		if err != nil {
			return "", "", err
		}
		var installID, appID string
		if row.InstallID != nil {
			installID = *row.InstallID
		}
		if row.AppID != nil {
			appID = *row.AppID
		}
		return installID, appID, nil
	}
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

// resolveRunnerOwner authorizes a runner via its group's owner: install
// runners inherit their install's chain, org runners resolve to the bare org
// tier (org-wide permission required). This is a fixed single hop — runner
// groups are owned directly by an install or the org, never by another child.
func resolveRunnerOwner(ctx *gin.Context, db *gorm.DB, orgID, id string) (string, string, error) {
	var row struct {
		OwnerID   string
		OwnerType string
	}
	err := db.WithContext(ctx).
		Model(&app.Runner{}).
		Select("runner_groups.owner_id, runner_groups.owner_type").
		Joins("JOIN runner_groups ON runner_groups.id = runners.runner_group_id AND runner_groups.deleted_at = 0").
		Where("runners.org_id = ?", orgID).
		Where("runners.id = ?", id).
		Take(&row).Error
	if err != nil {
		return "", "", err
	}

	if row.OwnerType != "installs" && row.OwnerType != "install" {
		return "", "", nil
	}

	var inst struct{ AppID string }
	err = db.WithContext(ctx).
		Model(&app.Install{}).
		Select("app_id").
		Where("org_id = ?", orgID).
		Where("id = ?", row.OwnerID).
		Take(&inst).Error
	if err != nil {
		return "", "", err
	}
	return row.OwnerID, inst.AppID, nil
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
