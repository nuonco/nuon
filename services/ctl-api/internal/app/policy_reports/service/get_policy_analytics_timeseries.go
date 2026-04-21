package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type SeriesPoint struct {
	Labels      map[string]string `json:"labels"`
	Evaluations int               `json:"evaluations"`
	Denies      int               `json:"denies"`
	Warns       int               `json:"warns"`
	Passes      int               `json:"passes"`
}

type TimeseriesBucket struct {
	Time        time.Time      `json:"time"`
	Evaluations int            `json:"evaluations"`
	Denies      int            `json:"denies"`
	Warns       int            `json:"warns"`
	Passes      int            `json:"passes"`
	Series      []*SeriesPoint `json:"series,omitempty"`
}

type PolicyAnalyticsTimeseries struct {
	Interval string              `json:"interval"`
	GroupBy  []string            `json:"group_by"`
	Start    time.Time           `json:"start"`
	End      time.Time           `json:"end"`
	Buckets  []*TimeseriesBucket `json:"buckets"`
}

var validGroupByDimensions = map[string]struct{}{
	"policy_id":    {},
	"install_id":   {},
	"component_id": {},
	"owner_type":   {},
}

// @ID						GetPolicyAnalyticsTimeseries
// @Summary				get policy analytics timeseries
// @Description.markdown	get_policy_analytics_timeseries.md
// @Param					app_id		path	string	true	"app ID"
// @Param					start		query	string	false	"start time (RFC3339)"
// @Param					end			query	string	false	"end time (RFC3339)"
// @Param					group_by	query	string	false	"comma-separated dimensions: policy_id, install_id, component_id, owner_type"
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
// @Success				200	{object}	PolicyAnalyticsTimeseries
// @Router					/v1/apps/{app_id}/policy-analytics/timeseries [get]
func (s *service) GetPolicyAnalyticsTimeseries(ctx *gin.Context) {
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

	groupByDims, err := parseGroupBy(ctx.Query("group_by"))
	if err != nil {
		ctx.Error(stderr.ErrUser{Err: err, Description: err.Error()})
		return
	}

	installID := ctx.Query("install_id")
	policyID := ctx.Query("policy_id")

	ts, err := s.getPolicyAnalyticsTimeseries(ctx, org.ID, appID, start, end, groupByDims, installID, policyID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get policy analytics timeseries: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, ts)
}

func parseGroupBy(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}

	dims := strings.Split(raw, ",")
	seen := make(map[string]struct{}, len(dims))
	result := make([]string, 0, len(dims))

	for _, d := range dims {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		if _, ok := validGroupByDimensions[d]; !ok {
			return nil, fmt.Errorf("invalid group_by dimension: %s (valid: policy_id, install_id, component_id, owner_type)", d)
		}
		if _, dup := seen[d]; dup {
			continue
		}
		seen[d] = struct{}{}
		result = append(result, d)
	}

	return result, nil
}

type timeseriesRow struct {
	Bucket      time.Time `gorm:"column:bucket"`
	PolicyID    string    `gorm:"column:policy_id"`
	InstallID   string    `gorm:"column:install_id"`
	ComponentID string    `gorm:"column:component_id"`
	OwnerType   string    `gorm:"column:owner_type"`
	Evaluations int       `gorm:"column:evaluations"`
	Denies      int       `gorm:"column:denies"`
	Warns       int       `gorm:"column:warns"`
	Passes      int       `gorm:"column:passes"`
}

func (s *service) getPolicyAnalyticsTimeseries(ctx context.Context, orgID, appID string, start, end time.Time, groupByDims []string, installID, policyID string) (*PolicyAnalyticsTimeseries, error) {
	interval := intervalForRange(start, end)
	selectCols, groupCols, orderCols := buildTimeseriesSelectClauses(interval, groupByDims)

	whereClause, params := buildBaseWhereClause(analyticsFilter{
		OrgID: orgID, AppID: appID,
		Start: start, End: end,
		InstallID: installID, PolicyID: policyID,
	})

	query := fmt.Sprintf("SELECT %s FROM policy_report_events %s GROUP BY %s ORDER BY %s",
		selectCols, whereClause, groupCols, orderCols)

	var rows []timeseriesRow
	if err := s.chDB.WithContext(ctx).Raw(query, params...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("unable to query clickhouse: %w", err)
	}

	return &PolicyAnalyticsTimeseries{
		Interval: interval.Label,
		GroupBy:  groupByDims,
		Start:    start,
		End:      end,
		Buckets:  buildTimeseriesBuckets(rows, groupByDims),
	}, nil
}

func buildTimeseriesBuckets(rows []timeseriesRow, groupByDims []string) []*TimeseriesBucket {
	if len(groupByDims) == 0 {
		buckets := make([]*TimeseriesBucket, 0, len(rows))
		for _, row := range rows {
			buckets = append(buckets, &TimeseriesBucket{
				Time:        row.Bucket,
				Evaluations: row.Evaluations,
				Denies:      row.Denies,
				Warns:       row.Warns,
				Passes:      row.Passes,
			})
		}
		return buckets
	}

	bucketIndex := make(map[time.Time]*TimeseriesBucket)
	var bucketOrder []time.Time

	for _, row := range rows {
		b, exists := bucketIndex[row.Bucket]
		if !exists {
			b = &TimeseriesBucket{Time: row.Bucket}
			bucketIndex[row.Bucket] = b
			bucketOrder = append(bucketOrder, row.Bucket)
		}
		b.Evaluations += row.Evaluations
		b.Denies += row.Denies
		b.Warns += row.Warns
		b.Passes += row.Passes

		labels := extractLabels(row, groupByDims)
		b.Series = append(b.Series, &SeriesPoint{
			Labels:      labels,
			Evaluations: row.Evaluations,
			Denies:      row.Denies,
			Warns:       row.Warns,
			Passes:      row.Passes,
		})
	}

	buckets := make([]*TimeseriesBucket, 0, len(bucketOrder))
	for _, t := range bucketOrder {
		buckets = append(buckets, bucketIndex[t])
	}
	return buckets
}

func extractLabels(row timeseriesRow, dims []string) map[string]string {
	labels := make(map[string]string, len(dims))
	for _, dim := range dims {
		switch dim {
		case "policy_id":
			labels["policy_id"] = row.PolicyID
		case "install_id":
			labels["install_id"] = row.InstallID
		case "component_id":
			labels["component_id"] = row.ComponentID
		case "owner_type":
			labels["owner_type"] = row.OwnerType
		}
	}
	return labels
}
