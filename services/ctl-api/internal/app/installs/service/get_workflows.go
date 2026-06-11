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

	// Search matches the user-visible title each workflow is rendered with
	// (e.g. "Deploying to install (rds_cluster_temporal)"). The title lives
	// in the `name` column — a STORED generated column maintained by
	// Postgres, see migrations.Migration108InstallWorkflowsNameGenerated —
	// so we don't have to recompute it here. Whitespace tokens are AND'd so
	// a query like "deploying rds" matches a title containing both words in
	// any order. Workflow id is also accepted so users can paste a ULID.
	for _, token := range strings.Fields(search) {
		like := "%" + token + "%"
		query = query.Where("name ILIKE ? OR id ILIKE ?", like, like)
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
