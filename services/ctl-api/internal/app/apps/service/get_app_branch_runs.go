package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

// @ID						GetAppBranchRuns
// @Summary				get app branch workflow runs
// @Description			Returns workflow runs for an app branch ordered by creation time (descending)
// @Tags					apps
// @Param					app_id			path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					offset			query	int		false	"offset of results to return"	Default(0)
// @Param					limit			query	int		false	"limit of results to return"	Default(10)
// @Param					page			query	int		false	"page number of results to return"	Default(0)
// @Param					planonly		query	bool	false	"exclude preview (plan only) runs when set to false"	Default(true)
// @Param					preview			query	bool	false	"return only preview runs when true, only rollout runs when false"
// @Param					q				query	string	false	"case-insensitive substring match against run title and id"
// @Param					type			query	string	false	"filter by workflow type (comma-separated for several types)"
// @Param					status			query	string	false	"filter by workflow status (comma-separated for several statuses)"
// @Param					created_at_gte	query	string	false	"filter runs created after timestamp (RFC3339 format)"
// @Param					created_at_lte	query	string	false	"filter runs created before timestamp (RFC3339 format)"
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.Workflow
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/runs [get]
func (s *service) GetAppBranchRuns(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureAppBranches)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check feature: %w", err))
		return
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureAppBranches))
		return
	}

	appID := ctx.Param("app_id")
	appBranchID := ctx.Param("app_branch_id")

	planOnly := true
	if planOnlyParam := ctx.Query("planonly"); planOnlyParam != "" {
		planOnly, err = strconv.ParseBool(planOnlyParam)
		if err != nil {
			ctx.Error(fmt.Errorf("invalid planonly parameter: %w", err))
			return
		}
	}

	filters, err := parseAppBranchRunFilters(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Where(app.AppBranch{
			OrgID: org.ID,
			AppID: appID,
		}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
		return
	}

	workflows, err := s.getAppBranchRuns(ctx, appBranchID, planOnly, filters)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get workflows: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, workflows)
}

type appBranchRunFilters struct {
	q            string
	types        []string
	statuses     []string
	preview      *bool
	createdAtGte *time.Time
	createdAtLte *time.Time
}

func parseAppBranchRunFilters(ctx *gin.Context) (appBranchRunFilters, error) {
	filters := appBranchRunFilters{
		q:        ctx.Query("q"),
		types:    parseCommaSeparated(ctx.Query("type")),
		statuses: parseCommaSeparated(ctx.Query("status")),
	}

	if param := ctx.Query("preview"); param != "" {
		parsed, err := strconv.ParseBool(param)
		if err != nil {
			return filters, stderr.ErrUser{
				Err:         fmt.Errorf("invalid preview parameter: %w", err),
				Description: "preview must be true or false",
			}
		}
		filters.preview = &parsed
	}

	if param := ctx.Query("created_at_gte"); param != "" {
		parsed, err := parseRFC3339Param("created_at_gte", param)
		if err != nil {
			return filters, err
		}
		filters.createdAtGte = parsed
	}

	if param := ctx.Query("created_at_lte"); param != "" {
		parsed, err := parseRFC3339Param("created_at_lte", param)
		if err != nil {
			return filters, err
		}
		filters.createdAtLte = parsed
	}

	return filters, nil
}

func parseRFC3339Param(name, value string) (*time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("invalid %s parameter: %w", name, err),
			Description: fmt.Sprintf("%s must be in RFC3339 format", name),
		}
	}
	return &parsed, nil
}

func parseCommaSeparated(raw string) []string {
	var values []string
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func (s *service) getAppBranchRuns(ctx *gin.Context, appBranchID string, includePlanOnly bool, filters appBranchRunFilters) ([]app.Workflow, error) {
	var workflows []app.Workflow

	query := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("CreatedBy").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("group_idx, group_retry_idx, idx, created_at asc")
		}).
		Preload("Steps.CreatedBy").
		Preload("Steps.Approval").
		Preload("Steps.Approval.Response").
		Preload("AppBranchRuns").
		Preload("AppBranchRuns.AppBranchConfig").
		Preload("AppBranchRuns.AppBranchConfig.ConnectedGithubVCSConfig").
		Preload("AppBranchRuns.VCSConnectionCommit").
		Preload("AppBranchRuns.Comparison").
		Preload("AppBranchRuns.Comparison.BaseRun").
		Preload("AppBranchRuns.Comparison.BaseRun.VCSConnectionCommit").
		Preload("AppBranchRuns.Preview").
		Where("owner_type = ?", "app_branches").
		Where("owner_id = ?", appBranchID).
		Order("created_at DESC")

	if !includePlanOnly {
		query = query.Where("plan_only = ?", false)
	}

	if len(filters.types) > 0 {
		query = query.Where("type IN ?", filters.types)
	}

	if len(filters.statuses) > 0 {
		query = query.Where("status->>'status' IN ?", filters.statuses)
	}

	if filters.preview != nil {
		query = query.Where("plan_only = ?", *filters.preview)
	}

	for _, token := range strings.Fields(filters.q) {
		like := "%" + token + "%"
		query = query.Where("name ILIKE ? OR id ILIKE ?", like, like)
	}

	if filters.createdAtGte != nil {
		query = query.Where("created_at >= ?", filters.createdAtGte)
	}

	if filters.createdAtLte != nil {
		query = query.Where("created_at <= ?", filters.createdAtLte)
	}

	res := query.Find(&workflows)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get workflows: %w", res.Error)
	}

	workflows, err := db.HandlePaginatedResponse(ctx, workflows)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return workflows, nil
}
