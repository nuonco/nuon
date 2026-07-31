package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type InstallComponentHealthSummary struct {
	InstallComponentID string `json:"install_component_id"`
	// ComponentID is what dashboard component routes are keyed by — a link
	// built from the install-component id instead dead-ends on an empty page.
	ComponentID   string  `json:"component_id"`
	ComponentName string  `json:"component_name"`
	CurrentHealth string  `json:"current_health"`
	UptimePercent float64 `json:"uptime_percent"`
	// ObservedSeconds distinguishes "no data" from "0% up" — without it a
	// component that was never observed renders as total downtime.
	ObservedSeconds int64 `json:"observed_seconds"`
}

type InstallHealthTimelineResponse struct {
	InstallID       string                          `json:"install_id"`
	Days            int                             `json:"days"`
	CurrentHealth   string                          `json:"current_health"`
	UptimePercent   float64                         `json:"uptime_percent"`
	ObservedSeconds int64                           `json:"observed_seconds"`
	Daily           []dailyHealthBucket             `json:"daily"`
	Components      []InstallComponentHealthSummary `json:"components"`

	// ClusterAccessError is why health cannot currently inspect the install's
	// cluster, empty when it can. Surfaced once here rather than per component.
	ClusterAccessError string `json:"cluster_access_error,omitzero"`
}

// @ID						GetInstallHealthTimeline
// @Summary				install health timeline
// @Description			Returns the install's health history aggregated across its components: uptime_percent and observed_seconds are the worst component's, daily[].health is the worst verdict across components for that day, and components lists each component's own current health and uptime. Requires the component-health feature.
// @Param					install_id	path	string	true	"install ID"
// @Param					days		query	int		false	"size of the window in days, clamped to 1-90"	Default(90)
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	service.InstallHealthTimelineResponse
// @Router					/v1/installs/{install_id}/health/timeline [get]
func (s *service) GetInstallHealthTimeline(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	days, err := strconv.Atoi(ctx.DefaultQuery("days", strconv.Itoa(healthTimelineDefaultDays)))
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid days query param: %w", err),
			Description: "days must be an integer",
		})
		return
	}
	days = clampHealthTimelineDays(days)

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireComponentHealthFeature(ctx, org); err != nil {
		ctx.Error(err)
		return
	}

	resp, err := s.getInstallHealthTimeline(ctx, org.ID, installID, days)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install health timeline: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getInstallHealthTimeline(ctx context.Context, orgID, installID string, days int) (*InstallHealthTimelineResponse, error) {
	var comps []app.InstallComponent
	if err := s.db.WithContext(ctx).
		Preload("Component").
		Where(app.InstallComponent{InstallID: installID, OrgID: orgID}).
		Find(&comps).Error; err != nil {
		return nil, fmt.Errorf("unable to list install components: %w", err)
	}

	// windowFrom anchors the calendar-day buckets; spanFrom clamps to the
	// health baseline for uptime. Folding the daily grid from spanFrom instead
	// would shift bucket boundaries by the baseline's time of day.
	windowFrom, to := healthWindow(time.Now(), days)
	baseline, err := s.healthBaseline(ctx, orgID, installID)
	if err != nil {
		return nil, err
	}
	if baseline.IsZero() {
		// No explicit reset: start from the first verdict this install ever
		// produced, so enabling the feature doesn't read as 90 days of downtime.
		firstSeen, err := s.firstHealthObservedAt(ctx, orgID, installID)
		if err != nil {
			return nil, err
		}
		baseline = firstSeen
	}
	spanFrom := clampToBaseline(windowFrom, baseline)

	statuses := make([]app.InstallComponentHealthStatus, 0, len(comps))
	summaries := make([]InstallComponentHealthSummary, 0, len(comps))
	dailyPerComponent := make([][]dailyHealthBucket, 0, len(comps))

	worstFound := false
	var worstUptime float64
	var worstObserved int64

	for i := range comps {
		c := &comps[i]
		statuses = append(statuses, c.HealthStatus)

		transitions, err := s.listHealthTransitions(ctx, orgID, installID, c.ID, spanFrom, to)
		if err != nil {
			return nil, err
		}
		seed, err := s.healthAtWindowStart(ctx, orgID, installID, c.ID, spanFrom)
		if err != nil {
			return nil, err
		}
		spans := healthSpans(transitions, spanFrom, to, seed)
		totals := sumSpans(spans)
		uptime := totals.uptimePercent()

		summaries = append(summaries, InstallComponentHealthSummary{
			InstallComponentID: c.ID,
			ComponentID:        c.ComponentID,
			ComponentName:      c.Component.Name,
			CurrentHealth:      string(c.HealthStatus),
			UptimePercent:      uptime,
			ObservedSeconds:    totals.observedSeconds(),
		})
		dailyPerComponent = append(dailyPerComponent, foldDailyHealth(spans, windowFrom, days))

		if c.HealthStatus == app.InstallComponentHealthStatusUnset || c.HealthStatus == app.InstallComponentHealthStatusNotApplicable {
			continue
		}
		if !worstFound || uptime < worstUptime {
			worstFound = true
			worstUptime = uptime
			worstObserved = totals.observedSeconds()
		}
	}

	currentHealth, _ := app.CompositeComponentHealthStatus(statuses)

	// Best-effort: the cluster-access reason is context on top of the timeline,
	// never a reason to fail it. Blanking the whole health view because one
	// column could not be read would be far worse than omitting the banner.
	var install app.Install
	if err := s.db.WithContext(ctx).
		Select("id", "health_cluster_error").
		Where(app.Install{ID: installID, OrgID: orgID}).
		First(&install).Error; err != nil {
		s.l.Warn("unable to read install cluster access error",
			zap.String("install_id", installID), zap.Error(err))
	}

	resp := &InstallHealthTimelineResponse{
		InstallID:          installID,
		Days:               days,
		CurrentHealth:      string(currentHealth),
		Daily:              worstDailyAcrossComponents(dailyPerComponent, windowFrom, days),
		Components:         summaries,
		ClusterAccessError: install.HealthClusterError,
	}
	if worstFound {
		resp.UptimePercent = worstUptime
		resp.ObservedSeconds = worstObserved
	}

	return resp, nil
}
