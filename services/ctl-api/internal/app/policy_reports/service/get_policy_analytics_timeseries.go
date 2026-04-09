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

type TimeseriesBucket struct {
	Time        time.Time                   `json:"time"`
	Evaluations int                         `json:"evaluations"`
	Denies      int                         `json:"denies"`
	Warns       int                         `json:"warns"`
	Passes      int                         `json:"passes"`
	Groups      map[string]*TimeseriesGroup `json:"groups,omitempty"`
}

type TimeseriesGroup struct {
	Evaluations int `json:"evaluations"`
	Denies      int `json:"denies"`
	Warns       int `json:"warns"`
	Passes      int `json:"passes"`
}

type PolicyAnalyticsTimeseries struct {
	Interval string              `json:"interval"`
	Start    time.Time           `json:"start"`
	End      time.Time           `json:"end"`
	Buckets  []*TimeseriesBucket `json:"buckets"`
}

// @ID						GetPolicyAnalyticsTimeseries
// @Summary				get policy analytics timeseries
// @Description.markdown	get_policy_analytics_timeseries.md
// @Param					app_id		path	string	true	"app ID"
// @Param					start		query	string	false	"start time (RFC3339)"
// @Param					end			query	string	false	"end time (RFC3339)"
// @Param					group_by	query	string	false	"group by dimension: policy_id, install_id, component_id"
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

	groupBy := ctx.Query("group_by")
	if groupBy != "" && !isValidGroupBy(groupBy) {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid group_by: %s", groupBy),
			Description: "valid values: policy_id, install_id, component_id",
		})
		return
	}

	installID := ctx.Query("install_id")
	policyID := ctx.Query("policy_id")

	ts, err := s.getPolicyAnalyticsTimeseries(ctx, org.ID, appID, start, end, groupBy, installID, policyID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get policy analytics timeseries: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, ts)
}

func isValidGroupBy(groupBy string) bool {
	switch groupBy {
	case "policy_id", "install_id", "component_id":
		return true
	default:
		return false
	}
}

type timeseriesRow struct {
	Bucket      time.Time `gorm:"column:bucket"`
	GroupKey    string    `gorm:"column:group_key"`
	Evaluations int       `gorm:"column:evaluations"`
	Denies      int       `gorm:"column:denies"`
	Warns       int       `gorm:"column:warns"`
	Passes      int       `gorm:"column:passes"`
}

func (s *service) getPolicyAnalyticsTimeseries(ctx context.Context, orgID, appID string, start, end time.Time, groupBy, installID, policyID string) (*PolicyAnalyticsTimeseries, error) {
	interval := intervalForRange(start, end)
	selectCols, groupCols, orderCols := buildTimeseriesSelectClauses(interval, groupBy)

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
		Start:    start,
		End:      end,
		Buckets:  buildTimeseriesBuckets(rows, groupBy != ""),
	}, nil
}

func buildTimeseriesBuckets(rows []timeseriesRow, hasGroups bool) []*TimeseriesBucket {
	if !hasGroups {
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
			b = &TimeseriesBucket{
				Time:   row.Bucket,
				Groups: make(map[string]*TimeseriesGroup),
			}
			bucketIndex[row.Bucket] = b
			bucketOrder = append(bucketOrder, row.Bucket)
		}
		b.Evaluations += row.Evaluations
		b.Denies += row.Denies
		b.Warns += row.Warns
		b.Passes += row.Passes
		b.Groups[row.GroupKey] = &TimeseriesGroup{
			Evaluations: row.Evaluations,
			Denies:      row.Denies,
			Warns:       row.Warns,
			Passes:      row.Passes,
		}
	}

	buckets := make([]*TimeseriesBucket, 0, len(bucketOrder))
	for _, t := range bucketOrder {
		buckets = append(buckets, bucketIndex[t])
	}
	return buckets
}
