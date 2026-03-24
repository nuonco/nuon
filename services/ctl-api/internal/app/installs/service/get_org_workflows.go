package service

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetOrgWorkflows
// @Summary					get all workflows for the org
// @Description.markdown	get_org_workflows.md
// @Param					offset			query	int		false	"offset of results to return"	Default(0)
// @Param					limit			query	int		false	"limit of results to return"	Default(10)
// @Param					page			query	int		false	"page number of results to return"	Default(0)
// @Param					planonly		query	bool	false	"exclude plan only workflows when set to false"	Default(true)
// @Param					type			query	string	false	"filter by workflow type"
// @Param					finished		query	bool	false	"filter by finished state"
// @Param					created_at_gte	query	string	false	"filter workflows created after timestamp (RFC3339 format)"
// @Param					created_at_lte	query	string	false	"filter workflows created before timestamp (RFC3339 format)"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{array}		app.Workflow
// @Router					/v1/workflows [GET]
func (s *service) GetOrgWorkflows(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	planOnly := true
	planOnlyParam := ctx.Query("planonly")
	if planOnlyParam != "" {
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

	workflows, err := s.getOrgWorkflows(ctx, org.ID, planOnly, workflowType, finished, createdAtGte, createdAtLte)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org workflows"))
		return
	}

	ctx.JSON(http.StatusOK, workflows)
}

func (s *service) getOrgWorkflows(ctx *gin.Context, orgID string, excludePlanOnly bool, workflowType string, finished *bool, createdAtGte *time.Time, createdAtLte *time.Time) ([]app.Workflow, error) {
	var workflows []app.Workflow
	query := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("CreatedBy").
		Preload("Steps").
		Preload("Steps.CreatedBy").
		Preload("Steps.Approval").
		Preload("Steps.Approval.Response").
		Where("org_id = ?", orgID).
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

	if createdAtGte != nil {
		query = query.Where("created_at >= ?", createdAtGte)
	}

	if createdAtLte != nil {
		query = query.Where("created_at <= ?", createdAtLte)
	}

	res := query.Find(&workflows)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org workflows: %w", res.Error)
	}

	workflows, err := db.HandlePaginatedResponse(ctx, workflows)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return workflows, nil
}
