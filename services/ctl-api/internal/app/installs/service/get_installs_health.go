package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// InstallHealthSummary is one install's row in the fleet health rollup.
type InstallHealthSummary struct {
	InstallID           string `json:"install_id"`
	InstallName         string `json:"install_name"`
	AppID               string `json:"app_id"`
	Health              string `json:"health"`
	HealthDescription   string `json:"health_description"`
	UnhealthyComponents int    `json:"unhealthy_components"`
	DegradedComponents  int    `json:"degraded_components"`
}

// InstallsHealthResponse is the fleet-wide rollup a canary/bake controller
// polls. AllHealthy is true only when every counted install is Healthy —
// Unset counts in Total but not Healthy, so unevaluated installs can't pass.
type InstallsHealthResponse struct {
	Total      int                    `json:"total"`
	Healthy    int                    `json:"healthy"`
	Degraded   int                    `json:"degraded"`
	Unhealthy  int                    `json:"unhealthy"`
	Unknown    int                    `json:"unknown"`
	Unset      int                    `json:"unset"`
	AllHealthy bool                   `json:"all_healthy"`
	Installs   []InstallHealthSummary `json:"installs"`
}

// @ID						GetInstallsHealth
// @Summary				fleet health summary
// @Description			Returns the health rollup for every install the caller can see, optionally narrowed by app and by an install label selector. This is the primitive a canary or bake-period rollout polls to decide whether to continue: all_healthy is only true when every counted install is healthy, and installs whose health has never been evaluated are counted separately in unset rather than treated as a pass. Requires the component-health feature.
// @Param					app_id	query	string	false	"filter by app ID"
// @Param					labels	query	string	false	"label filter (key:value,key:value)"
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
// @Success				200	{object}	service.InstallsHealthResponse
// @Router					/v1/installs/health [get]
func (s *service) GetInstallsHealth(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireComponentHealthFeature(ctx, org); err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Query("app_id")
	lbls := labels.ParseLabelsQuery(ctx.Query("labels"))

	installs, err := s.getInstallsForHealthSummary(ctx, org.ID, appID, lbls)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get installs health summary: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, summarizeInstallsHealth(installs))
}

func (s *service) getInstallsForHealthSummary(ctx context.Context, orgID, appID string, lbls labels.Labels) ([]app.Install, error) {
	tx := s.db.WithContext(ctx).
		Scopes(labels.WithLabels("labels", lbls)).
		Preload("Org").
		Where(app.Install{OrgID: orgID}).
		Order("name ASC")

	if appID != "" {
		tx = tx.Where(app.Install{AppID: appID})
	}

	installs := make([]app.Install, 0)
	if err := tx.Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to query installs: %w", err)
	}

	return installs, nil
}

// summarizeInstallsHealth rolls installs (with AfterQuery-computed
// CompositeHealthStatus) into fleet counts; kept pure so it's unit-testable without a DB.
func summarizeInstallsHealth(installs []app.Install) *InstallsHealthResponse {
	resp := &InstallsHealthResponse{
		Installs: make([]InstallHealthSummary, 0, len(installs)),
	}

	for i := range installs {
		inst := &installs[i]
		unhealthy, degraded := countBadHealthComponents(inst.ComponentHealthStatuses)

		switch inst.CompositeHealthStatus {
		case app.InstallComponentHealthStatusHealthy:
			resp.Healthy++
		case app.InstallComponentHealthStatusDegraded:
			resp.Degraded++
		case app.InstallComponentHealthStatusUnhealthy:
			resp.Unhealthy++
		case app.InstallComponentHealthStatusUnset, app.InstallComponentHealthStatusNotApplicable:
			resp.Unset++
		default:
			// Progressing, Unknown, and any future non-terminal verdict roll up
			// here: not a pass, but distinct from "no data at all".
			resp.Unknown++
		}

		resp.Installs = append(resp.Installs, InstallHealthSummary{
			InstallID:           inst.ID,
			InstallName:         inst.Name,
			AppID:               inst.AppID,
			Health:              string(inst.CompositeHealthStatus),
			HealthDescription:   inst.CompositeHealthStatusDescription,
			UnhealthyComponents: unhealthy,
			DegradedComponents:  degraded,
		})
	}

	resp.Total = len(installs)
	resp.AllHealthy = resp.Total > 0 && resp.Healthy == resp.Total

	return resp
}

// countBadHealthComponents reads the install's denormalized per-component
// health hstore, avoiding a separate query per install in the fleet summary.
func countBadHealthComponents(statuses pgtype.Hstore) (unhealthy, degraded int) {
	for _, v := range statuses {
		if v == nil {
			continue
		}
		switch app.InstallComponentHealthStatus(*v) {
		case app.InstallComponentHealthStatusUnhealthy:
			unhealthy++
		case app.InstallComponentHealthStatusDegraded:
			degraded++
		}
	}
	return unhealthy, degraded
}
