package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// InstallDeploymentType is a normalized category for a deployment record.
type InstallDeploymentType string

const (
	InstallDeploymentTypeProvision           InstallDeploymentType = "provision"
	InstallDeploymentTypeReprovision         InstallDeploymentType = "reprovision"
	InstallDeploymentTypeSandboxReprovision  InstallDeploymentType = "sandbox_reprovision"
	InstallDeploymentTypeAppBranchUpdate     InstallDeploymentType = "app_branch_update"
	InstallDeploymentTypeComponentDeploy     InstallDeploymentType = "component_deploy"
	InstallDeploymentTypeImageUpdate         InstallDeploymentType = "image_update"
	InstallDeploymentTypeStackUpdate         InstallDeploymentType = "stack_update"
	InstallDeploymentTypeInstallConfigUpdate InstallDeploymentType = "install_config_update"
)

func workflowTypeToDeploymentType(t app.WorkflowType) InstallDeploymentType {
	switch t {
	case app.WorkflowTypeProvision:
		return InstallDeploymentTypeProvision
	case app.WorkflowTypeManualDeploy,
		app.WorkflowTypeDriftRun,
		app.WorkflowTypeDeployComponents,
		app.WorkflowTypeTeardownComponent,
		app.WorkflowTypeTeardownComponents,
		app.WorkflowTypeComponentEnabled,
		app.WorkflowTypeComponentDisabled,
		app.WorkflowTypeRecoverHelmRelease:
		return InstallDeploymentTypeComponentDeploy
	case app.WorkflowTypeReprovisionSandbox, app.WorkflowTypeDriftRunReprovisionSandbox:
		return InstallDeploymentTypeSandboxReprovision
	case app.WorkflowTypeReprovision:
		return InstallDeploymentTypeReprovision
	case app.WorkflowTypeReprovisionStack:
		return InstallDeploymentTypeStackUpdate
	case app.WorkflowTypeAppBranchConfigUpdate,
		app.WorkflowTypeAppBranchesRun,
		app.WorkflowTypeAppBranchesConfigRepoUpdate,
		app.WorkflowTypeAppBranchesComponentRepoUpdate,
		app.WorkflowTypeAppInstallSync:
		return InstallDeploymentTypeAppBranchUpdate
	case app.WorkflowTypeInputUpdate, app.WorkflowTypeSyncSecrets:
		return InstallDeploymentTypeInstallConfigUpdate
	default:
		return ""
	}
}

// InstallDeploymentWorkflowRef is a lightweight reference to the backing workflow.
type InstallDeploymentWorkflowRef struct {
	ID   string           `json:"id"`
	Type app.WorkflowType `json:"type"`
	Name string           `json:"name"`
}

// InstallDeploymentAppBranchRef captures the app branch that originated this change, when applicable.
type InstallDeploymentAppBranchRef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RunID     string `json:"run_id,omitempty"`
	GitRef    string `json:"git_ref,omitempty"`
	CommitSHA string `json:"sha,omitempty"`
}

// InstallDeploymentComponentRef describes the primary component affected by a single-component operation.
type InstallDeploymentConfigChange struct {
	Path          string `json:"path"`
	Operation     string `json:"operation"`
	PreviousValue string `json:"previous_value,omitempty"`
	NextValue     string `json:"next_value,omitempty"`
	IsRedacted    bool   `json:"is_redacted,omitempty"`
}

type InstallDeploymentChangeGroup struct {
	ID           string                          `json:"id"`
	Scope        string                          `json:"scope"`
	Label        string                          `json:"label"`
	ResourceName string                          `json:"resource_name,omitempty"`
	Summary      string                          `json:"summary"`
	Changes      []InstallDeploymentConfigChange `json:"changes"`
	FileDiff     string                          `json:"file_diff,omitempty"`
	DiffLanguage string                          `json:"diff_language,omitempty"`
}

type InstallDeploymentAffectedResources struct {
	Stack      bool     `json:"stack,omitempty"`
	Sandbox    bool     `json:"sandbox,omitempty"`
	Components []string `json:"components"`
	Images     []string `json:"images"`
}

// InstallDeployment is a normalized record representing a single change event on an install.
type InstallDeployment struct {
	ID        string                `json:"id"`
	Type      InstallDeploymentType `json:"type"`
	Status    string                `json:"status"`
	CreatedAt time.Time             `json:"created_at"`
	Title     string                `json:"title"`
	Summary   string                `json:"summary"`

	Workflow      *InstallDeploymentWorkflowRef  `json:"workflow,omitempty"`
	AppBranch     *InstallDeploymentAppBranchRef `json:"app_branch,omitempty"`
	ComponentName string                         `json:"component_name,omitempty"`

	AffectedResources InstallDeploymentAffectedResources `json:"affected_resources"`
	ChangeGroups      []InstallDeploymentChangeGroup     `json:"change_groups"`
}

// GetInstallDeploymentsResponse is the paginated response body for the deployments endpoint.
type GetInstallDeploymentsResponse struct {
	Deployments []InstallDeployment `json:"deployments"`
	Page        int                 `json:"page"`
	Offset      int                 `json:"offset"`
	Limit       int                 `json:"limit"`
	HasMore     bool                `json:"has_more"`
}

// @ID                    GetInstallDeployments
// @Summary               get normalized deployment feed for an install
// @Description.markdown  get_install_deployments.md
// @Param                 install_id      path   string  true   "install ID"
// @Param                 page            query  int     false  "page number"                                         Default(0)
// @Param                 offset          query  int     false  "offset of results to return"                         Default(0)
// @Param                 limit           query  int     false  "page size"                                           Default(20)
// @Param                 type            query  string  false  "filter by deployment type (comma-separated)"
// @Param                 status          query  string  false  "filter by workflow status (comma-separated)"
// @Param                 resource        query  string  false  "filter by affected stack, sandbox, or component name"
// @Param                 search          query  string  false  "case-insensitive substring match on id or title"
// @Param                 created_at_gte  query  string  false  "include deployments created at or after this RFC3339 timestamp"
// @Param                 created_at_lte  query  string  false  "include deployments created at or before this RFC3339 timestamp"
// @Tags                  installs
// @Accept                json
// @Produce               json
// @Security              APIKey
// @Security              OrgID
// @Failure               400  {object}  stderr.ErrResponse
// @Failure               401  {object}  stderr.ErrResponse
// @Failure               403  {object}  stderr.ErrResponse
// @Failure               404  {object}  stderr.ErrResponse
// @Failure               500  {object}  stderr.ErrResponse
// @Success               200  {object}  GetInstallDeploymentsResponse
// @Router                /v1/installs/{install_id}/deployments [GET]
func (s *service) GetInstallDeployments(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	limit := queryInt(ctx, "limit", 20, 1, 100)
	page := queryInt(ctx, "page", 0, 0, 10_000)
	offset := queryInt(ctx, "offset", 0, 0, 1_000_000)
	if ctx.Query("page") != "" {
		offset = page * limit
	} else {
		page = offset / limit
	}

	filterTypes := parseCommaSeparated(ctx.Query("type"))
	filterStatuses := parseCommaSeparated(ctx.Query("status"))
	resource := ctx.Query("resource")
	search := ctx.Query("search")

	var createdAtGte *time.Time
	if raw := ctx.Query("created_at_gte"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			ctx.Error(errors.Wrap(err, "invalid created_at_gte, must be RFC3339"))
			return
		}
		createdAtGte = &t
	}

	var createdAtLte *time.Time
	if raw := ctx.Query("created_at_lte"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			ctx.Error(errors.Wrap(err, "invalid created_at_lte, must be RFC3339"))
			return
		}
		createdAtLte = &t
	}

	resp, err := s.getInstallDeployments(ctx, org.ID, installID, page, offset, limit, filterTypes, filterStatuses, resource, search, createdAtGte, createdAtLte)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deployments: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getInstallDeployments(
	ctx *gin.Context,
	orgID, installID string,
	page, offset, limit int,
	filterTypes, filterStatuses []string,
	resource string,
	search string,
	createdAtGte, createdAtLte *time.Time,
) (*GetInstallDeploymentsResponse, error) {
	fetchLimit := offset + limit + 1
	workflowTypes := deploymentWorkflowTypes(filterTypes)
	if len(workflowTypes) == 0 {
		return &GetInstallDeploymentsResponse{
			Deployments: []InstallDeployment{},
			Page:        page,
			Offset:      offset,
			Limit:       limit,
			HasMore:     false,
		}, nil
	}

	query := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("InstallDeploys").
		Preload("InstallDeploys.InstallComponent").
		Preload("InstallDeploys.InstallComponent.Component").
		Preload("InstallDeploys.ComponentBuild").
		Where("owner_id = ?", installID).
		Where("org_id = ?", orgID).
		Where("plan_only = ?", false).
		Where("type IN ?", workflowTypes).
		Order("created_at DESC").
		Limit(fetchLimit)

	if len(filterStatuses) > 0 {
		query = query.Where("status->>'status' IN ?", filterStatuses)
	}

	switch resource {
	case "":
	case "stack":
		query = query.Where("type IN ?", []app.WorkflowType{
			app.WorkflowTypeProvision,
			app.WorkflowTypeReprovision,
			app.WorkflowTypeReprovisionStack,
			app.WorkflowTypeInputUpdate,
		})
	case "sandbox":
		query = query.Where("type IN ?", []app.WorkflowType{
			app.WorkflowTypeProvision,
			app.WorkflowTypeReprovision,
			app.WorkflowTypeReprovisionSandbox,
			app.WorkflowTypeDriftRunReprovisionSandbox,
			app.WorkflowTypeInputUpdate,
		})
	default:
		query = query.Where(`
			EXISTS (
				SELECT 1
				FROM install_deploys d
				JOIN install_components ic ON ic.id = d.install_component_id
				JOIN components c ON c.id = ic.component_id
				WHERE d.install_workflow_id = install_workflows.id
				  AND d.deleted_at = 0
				  AND ic.deleted_at = 0
				  AND c.deleted_at = 0
				  AND c.name = ?
			)`, resource)
	}

	for _, token := range strings.Fields(search) {
		like := "%" + token + "%"
		query = query.Where("name ILIKE ? OR id ILIKE ?", like, like)
	}

	if createdAtGte != nil {
		query = query.Where("created_at >= ?", createdAtGte)
	}
	if createdAtLte != nil {
		query = query.Where("created_at <= ?", createdAtLte)
	}

	var workflows []app.Workflow
	if err := query.Find(&workflows).Error; err != nil {
		return nil, fmt.Errorf("unable to query workflows: %w", err)
	}

	workflowIDs := make([]string, 0, len(workflows))
	for i := range workflows {
		workflowIDs = append(workflowIDs, workflows[i].ID)
	}
	versionsByWorkflowID := make(map[string]*app.InstallAppConfigVersion)
	diffsByWorkflowID := make(map[string]*app.InstallConfigDiff)
	if len(workflowIDs) > 0 {
		var versions []app.InstallAppConfigVersion
		if err := s.db.WithContext(ctx).
			Preload("AppBranchRun").
			Preload("AppBranchRun.AppBranch").
			Preload("AppBranchRun.VCSConnectionCommit").
			Where(app.InstallAppConfigVersion{OrgID: orgID, InstallID: installID}).
			Where("workflow_id IN ?", workflowIDs).
			Find(&versions).Error; err != nil {
			return nil, fmt.Errorf("unable to query install app config versions: %w", err)
		}
		blobCtx := blobstore.WithBlobService(ctx.Request.Context(), s.blobSvc)
		for i := range versions {
			if versions[i].WorkflowID != nil {
				versionsByWorkflowID[*versions[i].WorkflowID] = &versions[i]
				diff, err := loadInstallConfigDiff(blobCtx, &versions[i])
				if err != nil {
					return nil, fmt.Errorf("unable to load install app config diff: %w", err)
				}
				diffsByWorkflowID[*versions[i].WorkflowID] = diff
			}
		}
	}

	deployments := make([]InstallDeployment, 0, len(workflows))
	for i := range workflows {
		d := buildInstallDeployment(&workflows[i], versionsByWorkflowID[workflows[i].ID], diffsByWorkflowID[workflows[i].ID])
		if d.Type == "" {
			continue
		}
		deployments = append(deployments, d)
	}

	start := offset
	if start > len(deployments) {
		start = len(deployments)
	}
	end := start + limit
	hasMore := end < len(deployments)
	if end > len(deployments) {
		end = len(deployments)
	}

	return &GetInstallDeploymentsResponse{
		Deployments: deployments[start:end],
		Page:        page,
		Offset:      offset,
		Limit:       limit,
		HasMore:     hasMore,
	}, nil
}

func deploymentWorkflowTypes(filterTypes []string) []app.WorkflowType {
	filters := make(map[string]struct{}, len(filterTypes))
	for _, filterType := range filterTypes {
		filters[filterType] = struct{}{}
	}

	types := make([]app.WorkflowType, 0)
	for _, workflowType := range app.AllWorkflowTypes() {
		deploymentType := workflowTypeToDeploymentType(workflowType)
		if deploymentType == "" {
			continue
		}
		if len(filters) > 0 {
			if _, ok := filters[string(deploymentType)]; !ok {
				continue
			}
		}
		types = append(types, workflowType)
	}
	return types
}

func buildInstallDeployment(wf *app.Workflow, version *app.InstallAppConfigVersion, configDiff *app.InstallConfigDiff) InstallDeployment {
	status := wf.Status.Status
	if status == "" {
		status = "pending"
	}

	title := wf.Name
	if title == "" {
		title = wf.Type.Name()
	}

	d := InstallDeployment{
		ID:        wf.ID,
		Type:      workflowTypeToDeploymentType(wf.Type),
		Status:    string(status),
		CreatedAt: wf.CreatedAt,
		Title:     title,
		Summary:   wf.Type.Description(),
		Workflow: &InstallDeploymentWorkflowRef{
			ID:   wf.ID,
			Type: wf.Type,
			Name: wf.Name,
		},
		AffectedResources: InstallDeploymentAffectedResources{
			Components: []string{},
			Images:     []string{},
		},
		ChangeGroups: []InstallDeploymentChangeGroup{},
	}

	componentNames := collectComponentNames(wf.InstallDeploys)
	if len(componentNames) > 0 {
		d.AffectedResources.Components = componentNames
		d.ChangeGroups = append(d.ChangeGroups, InstallDeploymentChangeGroup{
			ID:      wf.ID + "-components",
			Scope:   "component",
			Label:   "Components",
			Summary: strings.Join(componentNames, ", "),
			Changes: []InstallDeploymentConfigChange{},
		})
	}

	if len(wf.InstallDeploys) == 1 {
		dep := &wf.InstallDeploys[0]
		d.ComponentName = dep.InstallComponent.Component.Name
	}

	if version != nil && version.AppBranchRunID != nil {
		run := &version.AppBranchRun
		branchRef := &InstallDeploymentAppBranchRef{
			ID:    run.AppBranchID,
			Name:  run.AppBranch.Name,
			RunID: run.ID,
		}
		meta := run.RunMetadata()
		branchRef.GitRef = meta.GitRef
		branchRef.CommitSHA = meta.HeadSHA
		if branchRef.CommitSHA == "" && run.VCSConnectionCommit != nil {
			branchRef.CommitSHA = run.VCSConnectionCommit.SHA
		}
		d.AppBranch = branchRef
	}

	categoryGroups := changeGroupsForWorkflowType(wf.Type)
	for _, cg := range categoryGroups {
		if cg.Scope == "component" && len(componentNames) > 0 {
			continue
		}
		d.ChangeGroups = append(d.ChangeGroups, cg)
		switch cg.Scope {
		case "stack":
			d.AffectedResources.Stack = true
		case "sandbox":
			d.AffectedResources.Sandbox = true
		}
	}
	applyInstallConfigDiff(&d, configDiff)

	return d
}

func collectComponentNames(deploys []app.InstallDeploy) []string {
	seen := make(map[string]struct{})
	names := make([]string, 0)
	for i := range deploys {
		name := deploys[i].InstallComponent.Component.Name
		if name == "" {
			name = deploys[i].InstallComponent.ComponentID
		}
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	return names
}

func changeGroupsForWorkflowType(t app.WorkflowType) []InstallDeploymentChangeGroup {
	switch t {
	case app.WorkflowTypeInputUpdate:
		return []InstallDeploymentChangeGroup{{ID: "inputs", Scope: "install_config", Label: "Inputs", Summary: "Install inputs changed", Changes: []InstallDeploymentConfigChange{}}}
	case app.WorkflowTypeReprovisionStack, app.WorkflowTypeReprovision:
		return []InstallDeploymentChangeGroup{{ID: "stack", Scope: "stack", Label: "Stack", Summary: "Install stack changed", Changes: []InstallDeploymentConfigChange{}}}
	case app.WorkflowTypeReprovisionSandbox, app.WorkflowTypeDriftRunReprovisionSandbox:
		return []InstallDeploymentChangeGroup{{ID: "sandbox", Scope: "sandbox", Label: "Sandbox", Summary: "Install sandbox changed", Changes: []InstallDeploymentConfigChange{}}}
	case app.WorkflowTypeProvision:
		return []InstallDeploymentChangeGroup{
			{ID: "stack", Scope: "stack", Label: "Stack", Summary: "Install stack provisioned", Changes: []InstallDeploymentConfigChange{}},
			{ID: "sandbox", Scope: "sandbox", Label: "Sandbox", Summary: "Install sandbox provisioned", Changes: []InstallDeploymentConfigChange{}},
			{ID: "components", Scope: "component", Label: "Components", Summary: "Install components deployed", Changes: []InstallDeploymentConfigChange{}},
		}
	default:
		return nil
	}
}

func applyInstallConfigDiff(deployment *InstallDeployment, configDiff *app.InstallConfigDiff) {
	if configDiff == nil {
		return
	}

	componentEntries := []struct {
		operation string
		entries   []app.ComponentDiffEntry
	}{
		{operation: "add", entries: configDiff.Added},
		{operation: "change", entries: configDiff.Changed},
		{operation: "remove", entries: configDiff.Removed},
	}
	for _, set := range componentEntries {
		for _, entry := range set.entries {
			name := entry.ComponentName
			if name == "" {
				name = entry.ComponentID
			}
			deployment.AffectedResources.Components = appendUniqueString(deployment.AffectedResources.Components, name)
			deployment.ChangeGroups = append(deployment.ChangeGroups, InstallDeploymentChangeGroup{
				ID:           deployment.ID + "-component-" + entry.ComponentID,
				Scope:        "component",
				Label:        "Component",
				ResourceName: name,
				Summary:      name,
				Changes: []InstallDeploymentConfigChange{{
					Path:          "build",
					Operation:     set.operation,
					PreviousValue: entry.OldBuildID,
					NextValue:     entry.NewBuildID,
				}},
			})
		}
	}

	if configDiff.StackChanged {
		deployment.AffectedResources.Stack = true
		deployment.ChangeGroups = append(deployment.ChangeGroups, InstallDeploymentChangeGroup{
			ID:      deployment.ID + "-stack",
			Scope:   "stack",
			Label:   "Stack",
			Summary: "Stack configuration changed",
			Changes: []InstallDeploymentConfigChange{{
				Path:          "config",
				Operation:     "change",
				PreviousValue: configDiff.StackOldID,
				NextValue:     configDiff.StackNewID,
			}},
		})
	}

	if configDiff.SandboxChanged || configDiff.SandboxBuildChanged {
		deployment.AffectedResources.Sandbox = true
		changes := make([]InstallDeploymentConfigChange, 0, 2)
		if configDiff.SandboxChanged {
			changes = append(changes, InstallDeploymentConfigChange{
				Path:          "config",
				Operation:     "change",
				PreviousValue: configDiff.SandboxOldID,
				NextValue:     configDiff.SandboxNewID,
			})
		}
		if configDiff.SandboxBuildChanged {
			changes = append(changes, InstallDeploymentConfigChange{
				Path:          "build",
				Operation:     "change",
				PreviousValue: configDiff.SandboxBuildOldID,
				NextValue:     configDiff.SandboxBuildNewID,
			})
		}
		deployment.ChangeGroups = append(deployment.ChangeGroups, InstallDeploymentChangeGroup{
			ID:      deployment.ID + "-sandbox",
			Scope:   "sandbox",
			Label:   "Sandbox",
			Summary: "Sandbox configuration changed",
			Changes: changes,
		})
	}
}

func appendUniqueString(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
