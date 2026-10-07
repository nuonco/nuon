package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

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
	resource, search string,
	createdAtGte, createdAtLte *time.Time,
) (*GetInstallDeploymentsResponse, error) {
	queryTypes := filterTypes
	if containsString(filterTypes, string(InstallDeploymentTypeImageUpdate)) && !containsString(filterTypes, string(InstallDeploymentTypeComponentDeploy)) {
		queryTypes = append(append([]string{}, filterTypes...), string(InstallDeploymentTypeComponentDeploy))
	}
	workflowTypes := deploymentWorkflowTypes(queryTypes)
	resp := &GetInstallDeploymentsResponse{
		Deployments: []InstallDeployment{},
		Page:        page,
		Offset:      offset,
		Limit:       limit,
	}
	if len(workflowTypes) == 0 {
		return resp, nil
	}

	query := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_deploys.created_at ASC").Order("install_deploys.id ASC")
		}).
		Preload("InstallDeploys.InstallComponent").
		Preload("InstallDeploys.InstallComponent.Component").
		Preload("InstallDeploys.ComponentBuild").
		Where("owner_id = ?", installID).
		Where("org_id = ?", orgID).
		Where("plan_only = ?", false).
		Where("type IN ?", workflowTypes).
		Order("created_at DESC").
		Order("id DESC").
		Limit(offset + limit + 1)

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
				  AND (
				c.name = ?
				OR EXISTS (
					SELECT 1 FROM component_builds cb
					WHERE cb.id = d.component_build_id
					  AND cb.deleted_at = 0
					  AND (cb.source_image = ? OR cb.source_ref = ?)
				)
			  )
			)`, resource, resource, resource)
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
	deployments, err := s.buildInstallDeployments(ctx, orgID, installID, workflows)
	if err != nil {
		return nil, err
	}
	if len(filterTypes) > 0 {
		filtered := make([]InstallDeployment, 0, len(deployments))
		for _, deployment := range deployments {
			if containsString(filterTypes, string(deployment.Type)) {
				filtered = append(filtered, deployment)
			}
		}
		deployments = filtered
	}

	start := min(offset, len(deployments))
	end := start + limit
	resp.HasMore = end < len(deployments)
	resp.Deployments = deployments[start:min(end, len(deployments))]
	return resp, nil
}
