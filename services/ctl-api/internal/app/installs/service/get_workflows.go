package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetWorkflows
// @Summary					get workflows
// @Description.markdown	get_workflows.md
// @Param					install_id					path	string	true	"install ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Param					planonly					query	bool	false	"exclude plan only workflows when set to false"	Default(true)
// @Param					type						query	string	false	"filter by workflow type"
// @Param					finished					query	bool	false	"filter by finished state"
// @Param					created_at_gte				query	string	false	"filter workflows created after timestamp (RFC3339 format)"
// @Param					created_at_lte				query	string	false	"filter workflows created before timestamp (RFC3339 format)"
// @Param					search						query	string	false	"case-insensitive substring match against workflow id, type, and metadata (component / action / runbook name)"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{array}		app.Workflow
// @Router					/v1/installs/{install_id}/workflows [GET]
func (s *service) GetWorkflows(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	planOnly := true
	planOnlyParam := ctx.Query("planonly")
	if planOnlyParam != "" {
		var err error
		planOnly, err = strconv.ParseBool(planOnlyParam)
		if err != nil {
			ctx.Error(errors.Wrap(err, "invalid planonly parameter"))
			return
		}
	}

	workflowType := ctx.Query("type")

	var finished *bool
	finishedParam := ctx.Query("finished")
	if finishedParam != "" {
		f, err := strconv.ParseBool(finishedParam)
		if err != nil {
			ctx.Error(errors.Wrap(err, "invalid finished parameter"))
			return
		}
		finished = &f
	}

	var createdAtGte *time.Time
	createdAtGteParam := ctx.Query("created_at_gte")
	if createdAtGteParam != "" {
		parsedTime, err := time.Parse(time.RFC3339, createdAtGteParam)
		if err != nil {
			ctx.Error(errors.Wrap(err, "invalid created_at_gte parameter, must be in RFC3339 format"))
			return
		}
		createdAtGte = &parsedTime
	}

	var createdAtLte *time.Time
	createdAtLteParam := ctx.Query("created_at_lte")
	if createdAtLteParam != "" {
		parsedTime, err := time.Parse(time.RFC3339, createdAtLteParam)
		if err != nil {
			ctx.Error(errors.Wrap(err, "invalid created_at_lte parameter, must be in RFC3339 format"))
			return
		}
		createdAtLte = &parsedTime
	}

	search := ctx.Query("search")

	workflows, err := s.getWorkflows(ctx, installID, planOnly, workflowType, search, finished, createdAtGte, createdAtLte)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflows"))
		return
	}

	ctx.JSON(http.StatusOK, workflows)
}

func (s *service) getWorkflows(ctx *gin.Context, installID string, excludePlanOnly bool, workflowType, search string, finished *bool, createAtGte *time.Time, createdAtLte *time.Time) ([]app.Workflow, error) {
	var workflows []app.Workflow
	query := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("CreatedBy").
		Preload("Steps").
		Preload("Steps.CreatedBy").
		Preload("Steps.Approval").
		Preload("Steps.Approval.Response").
		Where("owner_id = ?", installID).
		Order("created_at desc")

	if !excludePlanOnly {
		query = query.Where("plan_only = ?", false)
	}

	if finished != nil {
		if *finished {
			query = query.Where("finished_at IS NOT NULL")
		} else {
			query = query.Where("finished_at IS NULL")
		}
	}

	if workflowType != "" {
		query = query.Where("type = ?", workflowType)
	}

	// Search matches the user-visible title each workflow is rendered with in
	// the dashboard (e.g. "Deploying to install (rds_cluster_temporal)",
	// "Runbook run (deploy_control_plane)"). The title is computed inline in
	// SQL — see workflowTitleExpr — and must stay in sync with the Go
	// implementation in app.Workflow.AfterQuery, which produces the same
	// string for the API response. Whitespace tokens are AND'd so a query
	// like "deploying rds" matches a title containing both words in any
	// order. Workflow id is also accepted as a direct prefix match for the
	// case where users paste a ULID.
	for _, token := range strings.Fields(search) {
		like := "%" + token + "%"
		query = query.Where(workflowTitleExpr+" ILIKE ? OR id ILIKE ?", like, like)
	}

	if createAtGte != nil {
		query = query.Where("created_at >= ?", createAtGte)
	}

	if createdAtLte != nil {
		query = query.Where("created_at <= ?", createdAtLte)
	}

	res := query.Find(&workflows)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get workflow runs: %w", res.Error)
	}

	workflows, err := db.HandlePaginatedResponse(ctx, workflows)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return workflows, nil
}

// workflowTitleExpr is a SQL expression that reproduces the same humanized
// workflow title the dashboard renders. It mirrors app.Workflow.AfterQuery:
//
//   - Base label depends on workflow type and whether finished_at is set
//     (in-progress vs. past-tense form).
//   - Adhoc action runs get a dedicated title.
//   - The label is suffixed with " (workflow-name-suffix)" if present, then
//     additionally with the action workflow name or runbook name for action
//     and runbook workflows respectively.
//
// Used for search so that "Deploying to install (rds_cluster_temporal)" or
// "Running runbook (deploy_control_plane)" can be substring-matched against
// the exact text the user sees.
//
// IMPORTANT: keep this in sync with WorkflowType.Name / PastTenseName and
// Workflow.AfterQuery. New workflow types added there must be added here too,
// otherwise their workflows will be unsearchable by humanized title.
const workflowTitleExpr = `(
	CASE
	WHEN type = 'action_workflow_run' AND COALESCE(metadata->'adhoc_action', '') <> ''
		THEN 'Adhoc action run (' || COALESCE(metadata->'install_action_workflow_name', '') || ')'
	ELSE
		COALESCE(
			CASE
			WHEN finished_at IS NULL THEN
				CASE type
					WHEN 'provision' THEN 'Provisioning install'
					WHEN 'reprovision' THEN 'Reprovisioning install'
					WHEN 'deprovision' THEN 'Deprovisioning install'
					WHEN 'manual_deploy' THEN 'Deploying to install'
					WHEN 'drift_run' THEN 'Deploying to install'
					WHEN 'input_update' THEN 'Input Update'
					WHEN 'teardown_components' THEN 'Tearing down all components'
					WHEN 'deploy_components' THEN 'Deploying all components'
					WHEN 'reprovision_sandbox' THEN 'Reprovisioning sandbox'
					WHEN 'drift_run_reprovision_sandbox' THEN 'Reprovisioning sandbox'
					WHEN 'sync_secrets' THEN 'Syncing secrets'
					WHEN 'action_workflow_run' THEN 'Action run'
					WHEN 'app_config_build' THEN 'Building app config components'
					WHEN 'runbook_run' THEN 'Running runbook'
				END
			ELSE
				CASE type
					WHEN 'provision' THEN 'Provisioned install'
					WHEN 'reprovision' THEN 'Reprovisioned install'
					WHEN 'reprovision_sandbox' THEN 'Reprovisioned sandbox'
					WHEN 'drift_run_reprovision_sandbox' THEN 'Reprovisioned sandbox'
					WHEN 'deprovision' THEN 'Deprovisioned install'
					WHEN 'manual_deploy' THEN 'Deployed to install'
					WHEN 'drift_run' THEN 'Deployed to install'
					WHEN 'input_update' THEN 'Updated Input'
					WHEN 'teardown_components' THEN 'Tore down all components'
					WHEN 'deploy_components' THEN 'Deployed all components'
					WHEN 'sync_secrets' THEN 'Synced secrets'
					WHEN 'action_workflow_run' THEN 'Action run'
					WHEN 'app_config_build' THEN 'Built app config components'
					WHEN 'runbook_run' THEN 'Runbook run'
				END
			END,
			REPLACE(type, '_', ' ')
		)
		|| COALESCE(' (' || NULLIF(metadata->'workflow-name-suffix', '') || ')', '')
		|| CASE WHEN type = 'action_workflow_run'
				THEN COALESCE(' (' || NULLIF(metadata->'install_action_workflow_name', '') || ')', '')
				ELSE '' END
		|| CASE WHEN type = 'runbook_run'
				THEN COALESCE(' (' || NULLIF(metadata->'runbook_name', '') || ')', '')
				ELSE '' END
	END
)`
