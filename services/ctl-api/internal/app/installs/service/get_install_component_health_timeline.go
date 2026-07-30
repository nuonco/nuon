package service

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// HealthTransitionResponse is one recorded verdict change, newest first in
// the timeline responses.
type HealthTransitionResponse struct {
	FromHealth            string    `json:"from_health"`
	ToHealth              string    `json:"to_health"`
	Message               string    `json:"message"`
	RootResourceKind      string    `json:"root_resource_kind"`
	RootResourceNamespace string    `json:"root_resource_namespace"`
	RootResourceName      string    `json:"root_resource_name"`
	CorrelatedDeployID    string    `json:"correlated_deploy_id"`
	Diagnosis             string    `json:"diagnosis"`
	ObservedAt            time.Time `json:"observed_at"`
}

type InstallComponentHealthTimelineResponse struct {
	InstallComponentID string                     `json:"install_component_id"`
	Days               int                        `json:"days"`
	CurrentHealth      string                     `json:"current_health"`
	UptimePercent      float64                    `json:"uptime_percent"`
	ObservedSeconds    int64                      `json:"observed_seconds"`
	Transitions        []HealthTransitionResponse `json:"transitions"`
	Daily              []dailyHealthBucket        `json:"daily"`
}

// @ID						GetInstallComponentHealthTimeline
// @Summary				component health timeline
// @Description			Returns a component's health history over a window: recorded verdict transitions (newest first), daily worst-verdict buckets covering every day in the window, and an uptime percentage that excludes unknown time from both the numerator and denominator. Requires the component-health feature.
// @Param					install_id				path	string	true	"install ID"
// @Param					component_id	path	string	true	"component ID"
// @Param					days					query	int		false	"size of the window in days, clamped to 1-90"	Default(90)
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
// @Success				200	{object}	service.InstallComponentHealthTimelineResponse
// @Router					/v1/installs/{install_id}/components/{component_id}/health/timeline [get]
func (s *service) GetInstallComponentHealthTimeline(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

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

	resp, err := s.getInstallComponentHealthTimeline(ctx, org.ID, installID, componentID, days)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install component health timeline: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getInstallComponentHealthTimeline(ctx context.Context, orgID, installID, componentID string, days int) (*InstallComponentHealthTimelineResponse, error) {
	ic, err := s.findInstallComponent(ctx, orgID, installID, componentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install component: %w", err)
	}
	installComponentID := ic.ID

	// windowFrom anchors the calendar-day buckets; spanFrom clamps to the
	// health baseline — see getInstallHealthTimeline for why the two differ.
	windowFrom, to := healthWindow(time.Now(), days)
	baseline, err := s.healthBaseline(ctx, orgID, installID)
	if err != nil {
		return nil, err
	}
	spanFrom := clampToBaseline(windowFrom, baseline)

	// Transitions are history and ignore the baseline — a reset changes what
	// counts toward uptime, not what happened.
	transitions, err := s.listHealthTransitions(ctx, orgID, installID, installComponentID, windowFrom, to)
	if err != nil {
		return nil, err
	}

	seed, err := s.healthAtWindowStart(ctx, orgID, installID, ic.ID, spanFrom)
	if err != nil {
		return nil, err
	}
	spans := healthSpans(transitions, spanFrom, to, seed)
	totals := sumSpans(spans)

	sortedTransitions := make([]app.InstallComponentHealthTransition, len(transitions))
	copy(sortedTransitions, transitions)
	sort.Slice(sortedTransitions, func(i, j int) bool {
		return sortedTransitions[i].ObservedAt.After(sortedTransitions[j].ObservedAt)
	})

	txResp := make([]HealthTransitionResponse, 0, len(sortedTransitions))
	for _, t := range sortedTransitions {
		txResp = append(txResp, HealthTransitionResponse{
			FromHealth:            t.FromHealth,
			ToHealth:              t.ToHealth,
			Message:               t.Message,
			RootResourceKind:      t.RootResourceKind,
			RootResourceNamespace: t.RootResourceNamespace,
			RootResourceName:      t.RootResourceName,
			CorrelatedDeployID:    t.CorrelatedDeployID,
			Diagnosis:             t.Diagnosis,
			ObservedAt:            t.ObservedAt,
		})
	}

	return &InstallComponentHealthTimelineResponse{
		InstallComponentID: installComponentID,
		Days:               days,
		CurrentHealth:      string(ic.HealthStatus),
		UptimePercent:      totals.uptimePercent(),
		ObservedSeconds:    totals.observedSeconds(),
		Transitions:        txResp,
		Daily:              foldDailyHealth(spans, windowFrom, days),
	}, nil
}
