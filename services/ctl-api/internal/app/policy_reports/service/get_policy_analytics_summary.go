package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type PolicyAnalyticsSummary struct {
	TotalEvaluations int       `json:"total_evaluations"`
	TotalDenies      int       `json:"total_denies"`
	TotalWarns       int       `json:"total_warns"`
	TotalPasses      int       `json:"total_passes"`
	UniqueReports    int       `json:"unique_reports"`
	UniquePolicies   int       `json:"unique_policies"`
	Start            time.Time `json:"start"`
	End              time.Time `json:"end"`
}

// @ID						GetPolicyAnalyticsSummary
// @Summary				get policy analytics summary
// @Description.markdown	get_policy_analytics_summary.md
// @Param					app_id		path	string	true	"app ID"
// @Param					start		query	string	false	"start time (RFC3339)"
// @Param					end			query	string	false	"end time (RFC3339)"
// @Param					install_id	query	string	false	"filter by install ID"
// @Param					policy_id	query	string	false	"filter by policy ID"
// @Tags					policy-reports
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	PolicyAnalyticsSummary
// @Router					/v1/apps/{app_id}/policy-analytics/summary [get]
func (s *service) GetPolicyAnalyticsSummary(ctx *gin.Context) {
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

	installID := ctx.Query("install_id")
	policyID := ctx.Query("policy_id")

	summary, err := s.getPolicyAnalyticsSummary(ctx, org.ID, appID, start, end, installID, policyID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get policy analytics summary: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, summary)
}

func (s *service) getPolicyAnalyticsSummary(ctx context.Context, orgID, appID string, start, end time.Time, installID, policyID string) (*PolicyAnalyticsSummary, error) {
	whereClause, params := buildBaseWhereClause(analyticsFilter{
		OrgID: orgID, AppID: appID,
		Start: start, End: end,
		InstallID: installID, PolicyID: policyID,
	})

	query := fmt.Sprintf(`SELECT
			count()                       AS total_evaluations,
			countIf(outcome = 'deny')     AS total_denies,
			countIf(outcome = 'warn')     AS total_warns,
			countIf(outcome = 'pass')     AS total_passes,
			uniqExact(report_id)          AS unique_reports,
			uniqExact(policy_id)          AS unique_policies
		FROM policy_report_events
		%s`, whereClause)

	var result PolicyAnalyticsSummary
	if err := s.chDB.WithContext(ctx).Raw(query, params...).Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("unable to query clickhouse: %w", err)
	}

	result.Start = start
	result.End = end
	return &result, nil
}

func parseTimeRange(ctx *gin.Context) (time.Time, time.Time, error) {
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -30)

	if startStr := ctx.Query("start"); startStr != "" {
		parsed, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start time: %w", err)
		}
		start = parsed
	}

	if endStr := ctx.Query("end"); endStr != "" {
		parsed, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end time: %w", err)
		}
		end = parsed
	}

	if !start.Before(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("start must be before end")
	}

	return start, end, nil
}
