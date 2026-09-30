package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	resourceProviderKubernetes = "kubernetes"
	resourceProviderProbe      = "probe"
	resourceProviderAWS        = "aws"
	resourceProviderGCP        = "gcp"
	resourceProviderAzure      = "azure"

	installStatusStaleAfter = 5 * time.Minute
)

type InstallStatusResponse struct {
	Deployments  app.CompositeStatus `json:"deployments"`
	Resources    app.CompositeStatus `json:"resources"`
	HealthChecks app.CompositeStatus `json:"health_checks"`
}

// @ID						GetInstallStatus
// @Summary				install status
// @Description			Returns deployment, resource, and health-check status for an install. Each axis is a composite status with counts in metadata. Deployment state stays on the component lifecycle, so a failed deploy remains failed after the workload is repaired. Resources and health checks are the latest observations.
// @Param					install_id	path	string	true	"install ID"
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
// @Success				200	{object}	service.InstallStatusResponse
// @Router					/v1/installs/{install_id}/status [get]
func (s *service) GetInstallStatus(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	status, err := s.getInstallStatus(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install status: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, status)
}

func (s *service) getInstallStatus(ctx context.Context, orgID, installID string) (*InstallStatusResponse, error) {
	var install app.Install
	if err := s.db.WithContext(ctx).
		Where(app.Install{ID: installID, OrgID: orgID}).
		First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	var components []app.InstallComponent
	if err := s.db.WithContext(ctx).
		Where(app.InstallComponent{InstallID: install.ID, OrgID: orgID}).
		Find(&components).Error; err != nil {
		return nil, fmt.Errorf("unable to list install components: %w", err)
	}

	resources, err := s.getInstallResources(ctx, orgID, install.ID, installResourceFilters{})
	if err != nil {
		return nil, err
	}

	declared, err := s.declaredHealthChecksByInstallComponent(ctx, orgID, install.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	health := healthCheckStatus(resources, declared, deployedComponentIDs(components), install.HealthClusterError, now)
	checks, err := s.listInstallHealthchecks(ctx, orgID, install.ID, currentAppConfigID(&install))
	if err != nil {
		return nil, err
	}
	applyFailingHealthchecks(&health, failingHealthchecks(checks))

	resourcesStatus := resourceStatus(resources, install.HealthClusterError, now)
	drift, err := s.installConfigDrift(ctx, &install)
	if err != nil {
		return nil, err
	}
	if resourcesStatus.Metadata == nil {
		resourcesStatus.Metadata = map[string]any{}
	}
	resourcesStatus.Metadata["config_drift"] = drift.driftedCount()

	return &InstallStatusResponse{
		Deployments:  deploymentStatus(components, now),
		Resources:    resourcesStatus,
		HealthChecks: health,
	}, nil
}

func applyFailingHealthchecks(status *app.CompositeStatus, failed int) {
	if failed == 0 || status == nil {
		return
	}
	status.Status = app.Status(app.InstallComponentHealthStatusUnhealthy)
	status.StatusHumanDescription = "Unhealthy"
	counts, _ := status.Metadata["counts"].(map[string]int)
	if counts == nil {
		counts = map[string]int{}
		if status.Metadata == nil {
			status.Metadata = map[string]any{}
		}
		status.Metadata["counts"] = counts
	}
	counts["unhealthy"] += failed
}

func deployedComponentIDs(components []app.InstallComponent) map[string]bool {
	deployed := make(map[string]bool, len(components))
	for i := range components {
		deployed[components[i].ID] = components[i].Status != app.InstallComponentStatusDisabled && components[i].EverDeployed()
	}
	return deployed
}

func deploymentStatus(components []app.InstallComponent, now time.Time) app.CompositeStatus {
	var deployed, failed, progressing, notDeployed int
	for i := range components {
		switch components[i].Status {
		case app.InstallComponentStatusDisabled,
			app.InstallComponentStatusDeleted,
			app.InstallComponentStatusInactive:
			continue
		case app.InstallComponentStatusUnset:
			notDeployed++
		case app.InstallComponentStatusActive, app.InstallComponentStatusNoop:
			deployed++
		case app.InstallComponentStatusError, app.InstallComponentStatusDeleteFailed:
			failed++
		default:
			progressing++
		}
	}

	status := app.Status(app.InstallComponentStatusPending)
	desc := "No deployments"
	switch {
	case failed > 0:
		status = app.Status(app.InstallComponentStatusError)
		desc = countPhrase(failed, "deployment failed", "deployments failed")
	case progressing > 0:
		status = app.Status(app.InstallComponentStatusPending)
		desc = "In progress"
	case notDeployed > 0 && deployed > 0:
		status = app.Status(app.InstallComponentStatusPending)
		desc = fmt.Sprintf("%d of %d deployed", deployed, deployed+notDeployed)
	case notDeployed > 0:
		status = app.Status(app.InstallComponentStatusPending)
		desc = "Not deployed"
	case deployed > 0:
		status = app.Status(app.InstallComponentStatusActive)
		desc = "All deployed"
	}

	return app.CompositeStatus{
		CreatedAtTS:            now.Unix(),
		Status:                 status,
		StatusHumanDescription: desc,
		Metadata: map[string]any{
			"counts": map[string]int{
				"deployed":     deployed,
				"failed":       failed,
				"progressing":  progressing,
				"not_deployed": notDeployed,
			},
		},
	}
}

type healthCounts struct {
	healthy     int
	progressing int
	degraded    int
	unhealthy   int
	unknown     int
}

func (c *healthCounts) add(health string) {
	switch app.InstallComponentHealthStatus(health) {
	case app.InstallComponentHealthStatusHealthy:
		c.healthy++
	case app.InstallComponentHealthStatusProgressing:
		c.progressing++
	case app.InstallComponentHealthStatusDegraded:
		c.degraded++
	case app.InstallComponentHealthStatusUnhealthy:
		c.unhealthy++
	default:
		c.unknown++
	}
}

func (c healthCounts) asMap() map[string]int {
	return map[string]int{
		"healthy":     c.healthy,
		"progressing": c.progressing,
		"degraded":    c.degraded,
		"unhealthy":   c.unhealthy,
		"unknown":     c.unknown,
	}
}

func resourceStatus(resources []app.InstallComponentResourceState, clusterError string, now time.Time) app.CompositeStatus {
	var counts healthCounts
	for i := range resources {
		if !isStatusResource(resources[i]) {
			continue
		}
		health, ok := observationHealth(resources[i], clusterError, now)
		if !ok {
			continue
		}
		counts.add(health)
	}
	status := healthComposite(counts, "No resources", clusterError, now)
	if clusterError != "" && counts.unhealthy == 0 && counts.degraded == 0 && counts.healthy == 0 && counts.progressing == 0 {
		status.Status = app.Status(app.InstallComponentHealthStatusUnknown)
		status.StatusHumanDescription = "Cluster unavailable"
	}
	return status
}

func healthCheckStatus(
	resources []app.InstallComponentResourceState,
	declared map[string]map[string]bool,
	deployed map[string]bool,
	clusterError string,
	now time.Time,
) app.CompositeStatus {
	var counts healthCounts
	seen := map[string]map[string]bool{}
	for i := range resources {
		if !isHealthCheck(resources[i]) {
			continue
		}
		markSeen(seen, resources[i].InstallComponentID, resources[i].Name)
		health, ok := observationHealth(resources[i], clusterError, now)
		if !ok {
			continue
		}
		counts.add(health)
	}

	for installComponentID, names := range declared {
		if !deployed[installComponentID] {
			continue
		}
		for name := range names {
			if name == "" || seen[installComponentID][name] {
				continue
			}
			counts.unknown++
		}
	}

	return healthComposite(counts, "No health checks", clusterError, now)
}

func healthComposite(counts healthCounts, emptyDesc, clusterError string, now time.Time) app.CompositeStatus {
	status := app.InstallComponentHealthStatusUnknown
	desc := emptyDesc
	switch {
	case counts.unhealthy > 0:
		status = app.InstallComponentHealthStatusUnhealthy
		desc = "Unhealthy"
	case counts.degraded > 0:
		status = app.InstallComponentHealthStatusDegraded
		desc = "Degraded"
	case counts.unknown > 0:
		status = app.InstallComponentHealthStatusUnknown
		desc = "Unknown"
	case counts.progressing > 0:
		status = app.InstallComponentHealthStatusProgressing
		desc = "Progressing"
	case counts.healthy > 0:
		status = app.InstallComponentHealthStatusHealthy
		desc = "All healthy"
	}
	metadata := map[string]any{"counts": counts.asMap()}
	if clusterError != "" {
		metadata["cluster_access_error"] = clusterError
	}
	return app.CompositeStatus{
		CreatedAtTS:            now.Unix(),
		Status:                 app.Status(status),
		StatusHumanDescription: desc,
		Metadata:               metadata,
	}
}

func markSeen(seen map[string]map[string]bool, installComponentID, name string) {
	if name == "" {
		return
	}
	if seen[installComponentID] == nil {
		seen[installComponentID] = map[string]bool{}
	}
	seen[installComponentID][name] = true
}

func isHealthCheck(r app.InstallComponentResourceState) bool {
	if r.RemovedFromConfig {
		return false
	}
	if r.Provider == app.InstallComponentResourceProviderCustom || r.Provider == resourceProviderProbe {
		return true
	}
	return isProbeKind(r.Kind)
}

func isStatusResource(r app.InstallComponentResourceState) bool {
	if r.RemovedFromConfig || isHealthCheck(r) {
		return false
	}
	switch r.Provider {
	case resourceProviderAWS, resourceProviderGCP, resourceProviderAzure:
		return false
	default:
		return true
	}
}

func observationHealth(r app.InstallComponentResourceState, clusterError string, now time.Time) (string, bool) {
	window := installStatusStaleAfter
	if r.StaleAfterSeconds > 0 {
		window = time.Duration(r.StaleAfterSeconds) * time.Second
	}
	stale := r.ObservedAt.IsZero() || now.Sub(r.ObservedAt) > window
	health := app.InstallComponentHealthStatus(r.Health)

	if r.Provider == resourceProviderKubernetes && clusterError != "" {
		if !stale && health.IsBadHealth() {
			return r.Health, true
		}
		return string(app.InstallComponentHealthStatusUnknown), true
	}
	if stale {
		return string(app.InstallComponentHealthStatusUnknown), true
	}
	switch health {
	case app.InstallComponentHealthStatusNotApplicable, app.InstallComponentHealthStatusUnset:
		return "", false
	case app.InstallComponentHealthStatusHealthy,
		app.InstallComponentHealthStatusProgressing,
		app.InstallComponentHealthStatusDegraded,
		app.InstallComponentHealthStatusUnhealthy,
		app.InstallComponentHealthStatusUnknown:
		return r.Health, true
	default:
		return string(app.InstallComponentHealthStatusUnknown), true
	}
}

func countPhrase(n int, singular, plural string) string {
	if n == 1 {
		return "1 " + singular
	}
	return fmt.Sprintf("%d %s", n, plural)
}
