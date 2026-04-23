package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type BreakdownEntry struct {
	Key         string `json:"key" gorm:"column:dimension_key"`
	Evaluations int    `json:"evaluations" gorm:"column:evaluations"`
	Denies      int    `json:"denies" gorm:"column:denies"`
	Warns       int    `json:"warns" gorm:"column:warns"`
	Passes      int    `json:"passes" gorm:"column:passes"`
}

type PolicyAnalyticsBreakdown struct {
	Dimension string            `json:"dimension"`
	Start     time.Time         `json:"start"`
	End       time.Time         `json:"end"`
	Entries   []*BreakdownEntry `json:"entries"`
}

var validBreakdownDimensions = map[string]struct{}{
	"policy_id":  {},
	"install_id": {},
	"owner_type": {},
}

// @ID						GetPolicyAnalyticsBreakdown
// @Summary				get policy analytics breakdown by dimension
// @Description.markdown	get_policy_analytics_breakdown.md
// @Param					app_id		path	string	true	"app ID"
// @Param					dimension	query	string	true	"dimension to break down by: policy_id, install_id, owner_type"
// @Param					start		query	string	false	"start time (RFC3339)"
// @Param					end			query	string	false	"end time (RFC3339)"
// @Param					install_id	query	string	false	"filter by install ID"
// @Param					policy_id	query	string	false	"filter by policy ID"
// @Param					limit		query	int		false	"max entries to return (default 10)"
// @Tags					policy-reports
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	PolicyAnalyticsBreakdown
// @Router					/v1/apps/{app_id}/policy-analytics/breakdown [get]
func (s *service) GetPolicyAnalyticsBreakdown(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	start, end, err := parseTimeRange(ctx)
	if err != nil {
		ctx.Error(stderr.ErrUser{Err: err, Description: "invalid time range"})
		return
	}

	dimension := ctx.Query("dimension")
	if _, ok := validBreakdownDimensions[dimension]; !ok {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid dimension: %s", dimension),
			Description: "valid dimensions: policy_id, install_id, owner_type",
		})
		return
	}

	limit := 10
	if limitStr := ctx.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 {
			ctx.Error(stderr.ErrUser{Err: fmt.Errorf("invalid limit: %s", limitStr), Description: "limit must be a positive integer"})
			return
		}
		limit = parsed
	}

	installID := ctx.Query("install_id")
	policyID := ctx.Query("policy_id")

	breakdown, err := s.getPolicyAnalyticsBreakdown(ctx, org.ID, appID, start, end, dimension, limit, installID, policyID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get policy analytics breakdown: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, breakdown)
}

func (s *service) getPolicyAnalyticsBreakdown(ctx context.Context, orgID, appID string, start, end time.Time, dimension string, limit int, installID, policyID string) (*PolicyAnalyticsBreakdown, error) {
	whereClause, params := buildBaseWhereClause(analyticsFilter{
		OrgID: orgID, AppID: appID,
		Start: start, End: end,
		InstallID: installID, PolicyID: policyID,
	})

	query := fmt.Sprintf(`SELECT
			%s AS dimension_key,
			count()                   AS evaluations,
			countIf(outcome = 'deny') AS denies,
			countIf(outcome = 'warn') AS warns,
			countIf(outcome = 'pass') AS passes
		FROM policy_report_events
		%s
		GROUP BY dimension_key
		ORDER BY denies DESC, warns DESC, evaluations DESC
		LIMIT ?`,
		dimension, whereClause)

	params = append(params, limit)

	var entries []*BreakdownEntry
	if err := s.chDB.WithContext(ctx).Raw(query, params...).Scan(&entries).Error; err != nil {
		return nil, fmt.Errorf("unable to query clickhouse: %w", err)
	}

	return &PolicyAnalyticsBreakdown{
		Dimension: dimension,
		Start:     start,
		End:       end,
		Entries:   entries,
	}, nil
}
