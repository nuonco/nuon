package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type AdminRunnerDetails struct {
	*app.Runner

	// Type is the runner group's owner type: "install" or "org".
	Type app.RunnerGroupType `json:"type,omitempty"`
	// OwnerID is the runner group's owner id (install id for install runners,
	// org id for org runners).
	OwnerID string `json:"owner_id,omitempty"`
	// Platform is the runner's cloud platform (aws/gcp/azure) derived from
	// the runner group's settings.
	Platform app.CloudPlatform `json:"platform,omitempty"`
	// Image is the runner container image URL from runner group settings.
	Image string `json:"image,omitempty"`
	// Tag is the runner container image tag from runner group settings.
	Tag string `json:"tag,omitempty"`
	// Uptime is the time since the latest active runner process started.
	Uptime time.Duration `json:"uptime,omitempty" swaggertype:"primitive,integer"`

	OrgName   string        `json:"org_name"`
	OrgLabels labels.Labels `json:"org_labels,omitempty"`

	InstallID     string        `json:"install_id,omitempty"`
	InstallName   string        `json:"install_name,omitempty"`
	InstallLabels labels.Labels `json:"install_labels,omitempty"`

	LatestHeartBeat   *app.LatestRunnerHeartBeat `json:"latest_heart_beat,omitempty"`
	LatestHealthCheck *app.RunnerHealthCheck     `json:"latest_health_check,omitempty"`
	RecentProcesses   []app.RunnerProcess        `json:"recent_processes,omitempty"`
}

const (
	adminRunnerDetailsHeartBeatLookback   = 6 * time.Hour
	adminRunnerDetailsHealthCheckLookback = 6 * time.Hour
	adminRunnerDetailsRecentProcessLimit  = 5
)

// @ID			AdminListRunnersDetails
// @BasePath	/v1/runners
// @Summary	Return all runners with runner group settings, owner labels, latest heartbeat/health check, and recent processes
// @Description	Admin list of runners enriched with their runner group settings,
// @Description	the labels of the runner group owner (org labels for org-type
// @Description	runners, install labels for install-type runners), the latest
// @Description	heartbeat and health check, and the last 5 runner processes.
// @Description	The optional `status` query parameter filters by
// @Description	`status_v2->>'status'` and may be repeated to match any of
// @Description	several statuses (e.g. `?status=error&status=offline`).
// @Param			offset	query	int			false	"offset of results to return"	Default(0)
// @Param			limit	query	int			false	"limit of results to return"	Default(10)
// @Param			page	query	int			false	"page number of results to return"	Default(0)
// @Param			status	query	[]string	false	"filter by composite status (repeatable)"	collectionFormat(multi)
// @Tags			runners/admin
// @Security		AdminEmail
// @Accept			json
// @Produce		json
// @Success		200	{array}	AdminRunnerDetails
// @Router			/v1/runners/details [GET]
func (s *service) AdminListRunnersDetails(ctx *gin.Context) {
	statuses := ctx.QueryArray("status")

	runners, err := s.listRunnersDetails(ctx, statuses)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, runners)
}

func (s *service) listRunnersDetails(ctx *gin.Context, statuses []string) ([]*AdminRunnerDetails, error) {
	var runners []*app.Runner
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("RunnerGroup").
		Preload("RunnerGroup.Settings").
		Order("created_at desc")
	if len(statuses) > 0 {
		tx = tx.Where("status_v2->>'status' IN ?", statuses)
	}
	if err := tx.Find(&runners).Error; err != nil {
		return nil, fmt.Errorf("unable to list runner details: %w", err)
	}

	runners, err := db.HandlePaginatedResponse(ctx, runners)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	runnerIDs := make([]string, 0, len(runners))
	orgIDs := make([]string, 0, len(runners))
	installIDsByGroup := make(map[string]string)
	for _, r := range runners {
		runnerIDs = append(runnerIDs, r.ID)
		orgIDs = append(orgIDs, r.OrgID)
		if r.RunnerGroup.Type == app.RunnerGroupTypeInstall {
			installIDsByGroup[r.RunnerGroup.ID] = r.RunnerGroup.OwnerID
		}
	}

	orgsByID, err := s.fetchOrgsByID(ctx, orgIDs)
	if err != nil {
		return nil, err
	}

	installsByID, err := s.fetchInstallsByID(ctx, installIDsByGroup)
	if err != nil {
		return nil, err
	}

	heartBeatsByRunner, err := s.fetchLatestHeartBeatsByRunner(ctx, runnerIDs)
	if err != nil {
		return nil, err
	}

	healthChecksByRunner, err := s.fetchLatestHealthChecksByRunner(ctx, runnerIDs)
	if err != nil {
		return nil, err
	}

	processesByRunner, err := s.fetchRecentProcessesByRunner(ctx, runnerIDs)
	if err != nil {
		return nil, err
	}

	items := make([]*AdminRunnerDetails, 0, len(runners))
	for _, r := range runners {
		item := &AdminRunnerDetails{
			Runner:            r,
			Type:              r.RunnerGroup.Type,
			OwnerID:           r.RunnerGroup.OwnerID,
			Platform:          r.RunnerGroup.Settings.Platform,
			Image:             r.RunnerGroup.Settings.ContainerImageURL,
			Tag:               r.RunnerGroup.Settings.ContainerImageTag,
			Uptime:            latestProcessUptime(processesByRunner[r.ID]),
			LatestHeartBeat:   heartBeatsByRunner[r.ID],
			LatestHealthCheck: healthChecksByRunner[r.ID],
			RecentProcesses:   processesByRunner[r.ID],
		}

		if org, ok := orgsByID[r.OrgID]; ok {
			item.OrgName = org.Name
			item.OrgLabels = org.Labels
		}

		if r.RunnerGroup.Type == app.RunnerGroupTypeInstall {
			if install, ok := installsByID[r.RunnerGroup.OwnerID]; ok {
				item.InstallID = install.ID
				item.InstallName = install.Name
				item.InstallLabels = install.Labels
			}
		}

		items = append(items, item)
	}

	return items, nil
}

// latestProcessUptime returns the uptime of the most recently started process
// in the list, or 0 if none are available. Processes are passed in
// created_at-desc order, but each process's own Uptime is populated by GORM's
// AfterQuery hook.
func latestProcessUptime(processes []app.RunnerProcess) time.Duration {
	for i := range processes {
		if processes[i].Uptime > 0 {
			return processes[i].Uptime
		}
	}
	return 0
}

func (s *service) fetchOrgsByID(ctx context.Context, orgIDs []string) (map[string]*app.Org, error) {
	out := make(map[string]*app.Org)
	if len(orgIDs) == 0 {
		return out, nil
	}

	var orgs []*app.Org
	if err := s.db.WithContext(ctx).
		Where("id IN ?", orgIDs).
		Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch orgs: %w", err)
	}

	for _, o := range orgs {
		out[o.ID] = o
	}
	return out, nil
}

func (s *service) fetchInstallsByID(ctx context.Context, installIDsByGroup map[string]string) (map[string]*app.Install, error) {
	out := make(map[string]*app.Install)
	if len(installIDsByGroup) == 0 {
		return out, nil
	}

	ids := make([]string, 0, len(installIDsByGroup))
	for _, id := range installIDsByGroup {
		ids = append(ids, id)
	}

	var installs []*app.Install
	if err := s.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch installs: %w", err)
	}

	for _, i := range installs {
		out[i.ID] = i
	}
	return out, nil
}

func (s *service) fetchLatestHeartBeatsByRunner(ctx context.Context, runnerIDs []string) (map[string]*app.LatestRunnerHeartBeat, error) {
	out := make(map[string]*app.LatestRunnerHeartBeat)
	if len(runnerIDs) == 0 {
		return out, nil
	}

	cutoff := time.Now().Add(-adminRunnerDetailsHeartBeatLookback)

	var heartBeats []*app.LatestRunnerHeartBeat
	if err := s.chDB.WithContext(ctx).
		Where("runner_id IN ?", runnerIDs).
		Where("created_at_latest > ?", cutoff).
		Find(&heartBeats).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch latest heart beats: %w", err)
	}

	for _, hb := range heartBeats {
		existing, ok := out[hb.RunnerID]
		if !ok || hb.CreatedAt.After(existing.CreatedAt) {
			out[hb.RunnerID] = hb
		}
	}
	return out, nil
}

func (s *service) fetchLatestHealthChecksByRunner(ctx context.Context, runnerIDs []string) (map[string]*app.RunnerHealthCheck, error) {
	out := make(map[string]*app.RunnerHealthCheck)
	if len(runnerIDs) == 0 {
		return out, nil
	}

	cutoff := time.Now().Add(-adminRunnerDetailsHealthCheckLookback)

	var healthChecks []*app.RunnerHealthCheck
	if err := s.chDB.WithContext(ctx).
		Scopes(scopes.WithOverrideTable("runner_health_checks_view_v2")).
		Where("runner_id IN ?", runnerIDs).
		Where("created_at > ?", cutoff).
		Order("created_at asc").
		Find(&healthChecks).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch latest health checks: %w", err)
	}

	for _, hc := range healthChecks {
		out[hc.RunnerID] = hc
	}
	return out, nil
}

func (s *service) fetchRecentProcessesByRunner(ctx context.Context, runnerIDs []string) (map[string][]app.RunnerProcess, error) {
	out := make(map[string][]app.RunnerProcess)
	if len(runnerIDs) == 0 {
		return out, nil
	}

	for _, runnerID := range runnerIDs {
		var processes []app.RunnerProcess
		if err := s.db.WithContext(ctx).
			Where("runner_id = ?", runnerID).
			Order("created_at desc").
			Limit(adminRunnerDetailsRecentProcessLimit).
			Find(&processes).Error; err != nil {
			return nil, fmt.Errorf("unable to fetch recent processes for runner %s: %w", runnerID, err)
		}
		if len(processes) > 0 {
			out[runnerID] = processes
		}
	}
	return out, nil
}
