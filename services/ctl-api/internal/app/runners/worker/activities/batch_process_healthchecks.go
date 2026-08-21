package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"go.temporal.io/sdk/activity"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const defaultBatchProcessLimit = 500

type BatchProcessHealthchecksRequest struct {
	OrgID    string `validate:"required"`
	CursorID string `json:"cursor_id"`
	Limit    int    `json:"limit"`
}

type BatchProcessHealthchecksResponse struct {
	NextCursorID string `json:"next_cursor_id"`
	Done         bool   `json:"done"`
	Checked      int    `json:"checked"`
	Active       int    `json:"active"`
	Offline      int    `json:"offline"`
	Inactive     int    `json:"inactive"`
	Shutdowns    int    `json:"shutdowns"`
	NoHeartbeat  int    `json:"no_heartbeat"`
	MissingQueue int    `json:"missing_queue"`
	Skipped      int    `json:"skipped"`
	Errors       int    `json:"errors"`
}

type pageHeartbeat struct {
	RunnerID  string    `gorm:"column:runner_id"`
	ProcessID string    `gorm:"column:process_id"`
	Version   string    `gorm:"column:latest_version"`
	CreatedAt time.Time `gorm:"column:latest_created_at"`
}

// BatchProcessHealthchecks checks one keyset page of an org's runner processes
// inline, replacing the per-process process_healthcheck signal fan-out.
//
// @temporal-gen-v2 activity
// @by-field OrgID
// @start-to-close-timeout 5m
// @heartbeat-timeout 60s
func (a *Activities) BatchProcessHealthchecks(ctx context.Context, req BatchProcessHealthchecksRequest) (*BatchProcessHealthchecksResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = defaultBatchProcessLimit
	}
	resp := &BatchProcessHealthchecksResponse{}

	var processes []app.RunnerProcess
	if res := a.db.WithContext(ctx).
		Select("id", "runner_id", "org_id", "created_by_id", "type", "composite_status", "started_at", "initial_health_check").
		Where("org_id = ?", req.OrgID).
		Where("composite_status->>'status' IN ?", []string{
			string(app.RunnerProcessStatusActive),
			string(app.RunnerProcessStatusOffline),
		}).
		Where("id > ?", req.CursorID).
		Order("id").
		Limit(limit).
		Find(&processes); res.Error != nil {
		return nil, fmt.Errorf("unable to list runner processes for org %s: %w", req.OrgID, res.Error)
	}

	resp.Done = len(processes) < limit
	if len(processes) == 0 {
		return resp, nil
	}
	resp.NextCursorID = processes[len(processes)-1].ID

	page, err := a.loadProcessPageContext(ctx, processes)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	healthRows := make([]app.RunnerHealthCheck, 0, len(processes))
	initialCheckIDs := make([]string, 0)

	for i := range processes {
		if i%batchHeartbeatEvery == 0 {
			activity.RecordHeartbeat(ctx, processes[i].ID)
		}
		p := processes[i]
		resp.Checked++
		entityStart := time.Now()

		runner, hasRunner := page.runnersByID[p.RunnerID]
		tags := a.processHealthTags(&p, runner, hasRunner, page)

		queue, hasQueue := page.queuesByName[fmt.Sprintf("runner-process-%s", p.ID)]
		if !hasQueue {
			resp.MissingQueue++
			continue
		}

		hb, hasHB := page.heartbeatsByProcess[p.ID]
		var hbAt *time.Time
		if hasHB {
			hbAt = &hb.CreatedAt
			tags["runner_version"] = hb.Version
		}

		ectx := cctx.SetOrgIDContext(ctx, p.OrgID)
		ectx = cctx.SetAccountIDContext(ectx, p.CreatedByID)

		action := decideProcessHealth(now, &p, hbAt)
		switch action {
		case processActionNoop:
			if !hasHB {
				resp.NoHeartbeat++
			} else {
				resp.Skipped++
			}
		case processActionShutdown:
			a.handleBatchProcessShutdown(ectx, &p, now, resp)
		case processActionInactive:
			a.handleBatchProcessInactive(ectx, &p, queue, now, tags, resp)
		case processActionOffline:
			if row, ok := a.handleBatchProcessOffline(ectx, &p, now, resp); ok {
				healthRows = append(healthRows, row)
			}
		case processActionActive:
			row, ok := a.handleBatchProcessActive(ectx, &p, now, resp)
			if !ok {
				break
			}
			healthRows = append(healthRows, row)
			a.checkBatchVersionMismatch(ectx, &p, runner, hasRunner, hb.Version, tags)
		}

		if hasHB && !p.InitialHealthCheck &&
			(action == processActionActive || action == processActionOffline || action == processActionInactive) {
			initialCheckIDs = append(initialCheckIDs, p.ID)
		}

		a.mw.Timing("runner.health_check.latency", time.Since(entityStart), metrics.ToTags(tags))
	}

	if len(healthRows) > 0 {
		if res := a.chDB.WithContext(ctx).CreateInBatches(&healthRows, 500); res.Error != nil {
			a.l.Warn("unable to insert batch process health checks",
				zap.String("org_id", req.OrgID), zap.Int("rows", len(healthRows)), zap.Error(res.Error))
		}
	}

	if len(initialCheckIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Model(&app.RunnerProcess{}).
			Where("id IN ?", initialCheckIDs).
			Update("initial_health_check", true); res.Error != nil {
			a.l.Warn("unable to mark initial health checks",
				zap.String("org_id", req.OrgID), zap.Error(res.Error))
		}
	}

	return resp, nil
}

type processPageContext struct {
	runnersByID         map[string]app.Runner
	installsByID        map[string]app.Install
	queuesByName        map[string]app.Queue
	heartbeatsByProcess map[string]pageHeartbeat
}

func (a *Activities) loadProcessPageContext(ctx context.Context, processes []app.RunnerProcess) (*processPageContext, error) {
	page := &processPageContext{
		runnersByID:         map[string]app.Runner{},
		installsByID:        map[string]app.Install{},
		queuesByName:        map[string]app.Queue{},
		heartbeatsByProcess: map[string]pageHeartbeat{},
	}

	runnerIDSet := map[string]struct{}{}
	processIDs := make([]string, 0, len(processes))
	queueNames := make([]string, 0, len(processes))
	for _, p := range processes {
		runnerIDSet[p.RunnerID] = struct{}{}
		processIDs = append(processIDs, p.ID)
		queueNames = append(queueNames, fmt.Sprintf("runner-process-%s", p.ID))
	}
	runnerIDs := make([]string, 0, len(runnerIDSet))
	for id := range runnerIDSet {
		runnerIDs = append(runnerIDs, id)
	}

	var runners []app.Runner
	if res := a.db.WithContext(ctx).
		Select("id", "org_id", "created_by_id", "status", "status_v2", "runner_group_id").
		Preload("Org").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Settings").
		Where("id IN ?", runnerIDs).
		Find(&runners); res.Error != nil {
		return nil, fmt.Errorf("unable to load runners for process page: %w", res.Error)
	}
	installIDs := make([]string, 0)
	for _, r := range runners {
		page.runnersByID[r.ID] = r
		if r.RunnerGroup.OwnerType == "installs" {
			installIDs = append(installIDs, r.RunnerGroup.OwnerID)
		}
	}

	if len(installIDs) > 0 {
		var installs []app.Install
		if res := a.db.WithContext(ctx).
			Where("id IN ?", installIDs).
			Find(&installs); res.Error != nil {
			return nil, fmt.Errorf("unable to load installs for process page: %w", res.Error)
		}
		for _, in := range installs {
			page.installsByID[in.ID] = in
		}
	}

	var queues []app.Queue
	if res := a.db.WithContext(ctx).
		Select("id", "name").
		Where("owner_type = ? AND name IN ?", "runners", queueNames).
		Find(&queues); res.Error != nil {
		return nil, fmt.Errorf("unable to load process queues for page: %w", res.Error)
	}
	for _, q := range queues {
		page.queuesByName[q.Name] = q
	}

	var heartbeats []pageHeartbeat
	if res := a.chDB.WithContext(ctx).Raw(`
		SELECT runner_id, process_id,
		       argMax(version, created_at) AS latest_version,
		       max(created_at)             AS latest_created_at
		FROM runner_heart_beats
		WHERE runner_id IN ? AND process_id IN ? AND deleted_at = 0
		GROUP BY runner_id, process_id`, runnerIDs, processIDs).
		Scan(&heartbeats); res.Error != nil {
		return nil, fmt.Errorf("unable to load heartbeats for process page: %w", res.Error)
	}
	for _, hb := range heartbeats {
		page.heartbeatsByProcess[hb.ProcessID] = hb
	}

	return page, nil
}

func (a *Activities) processHealthTags(p *app.RunnerProcess, runner app.Runner, hasRunner bool, page *processPageContext) map[string]string {
	tags := map[string]string{
		"runner_id":    p.RunnerID,
		"process_type": string(p.Type),
	}
	if !hasRunner {
		return tags
	}
	addLabels := func(labels map[string]string) {
		for k, v := range labels {
			if _, ok := tags[k]; !ok {
				tags[k] = v
			}
		}
	}
	tags["runner_type"] = string(runner.RunnerGroup.Type)
	tags["runner_status"] = string(runner.Status)
	tags["org_id"] = runner.OrgID
	tags["org_name"] = runner.Org.Name
	switch runner.RunnerGroup.OwnerType {
	case "installs":
		tags["install_id"] = runner.RunnerGroup.OwnerID
		if install, ok := page.installsByID[runner.RunnerGroup.OwnerID]; ok {
			tags["install_name"] = install.Name
			addLabels(install.Labels)
		}
	case "orgs":
		addLabels(runner.Org.Labels)
	}
	return tags
}

func (a *Activities) handleBatchProcessShutdown(ectx context.Context, p *app.RunnerProcess, now time.Time, resp *BatchProcessHealthchecksResponse) {
	// Without the per-process queue's MaxInFlight=1 serialization this can race
	// trigger_shutdown; an existing requested graceful shutdown absorbs the flag.
	var existing int64
	if res := a.db.WithContext(ectx).
		Model(&app.RunnerProcessShutdown{}).
		Where(app.RunnerProcessShutdown{
			RunnerProcessID: p.ID,
			Type:            app.RunnerProcessShutdownTypeGraceful,
		}).
		Where("composite_status->>'status' = ?", string(app.RunnerProcessShutdownStatusRequested)).
		Count(&existing); res.Error != nil {
		resp.Errors++
		a.l.Warn("unable to check for existing process shutdown", zap.String("process_id", p.ID), zap.Error(res.Error))
		return
	}

	if existing == 0 {
		if _, err := a.CreateRunnerProcessShutdown(ectx, CreateRunnerProcessShutdownRequest{
			RunnerProcessID: p.ID,
			Type:            app.RunnerProcessShutdownTypeGraceful,
			CompositeStatus: app.CompositeStatus{
				Status:                 app.Status(app.RunnerProcessShutdownStatusRequested),
				StatusHumanDescription: "shutdown requested via promotion",
				CreatedAtTS:            now.Unix(),
			},
		}); err != nil {
			resp.Errors++
			a.l.Warn("unable to create shutdown for process", zap.String("process_id", p.ID), zap.Error(err))
			return
		}
	}

	if err := a.ClearProcessShutdownRequested(ectx, ClearProcessShutdownRequestedRequest{ProcessID: p.ID}); err != nil {
		a.l.Warn("unable to clear shutdown_requested metadata", zap.String("process_id", p.ID), zap.Error(err))
	}
	resp.Shutdowns++
}

func (a *Activities) handleBatchProcessInactive(ectx context.Context, p *app.RunnerProcess, queue app.Queue, now time.Time, tags map[string]string, resp *BatchProcessHealthchecksResponse) {
	updated, err := a.guardedProcessStatusUpdate(ectx, p, app.RunnerProcessStatusInactive, "no heartbeat received for 5 minutes")
	if err != nil {
		resp.Errors++
		a.l.Warn("unable to update process status to inactive", zap.String("process_id", p.ID), zap.Error(err))
		return
	}
	if !updated {
		resp.Skipped++
		return
	}
	resp.Inactive++

	a.l.Warn("process inactive - no heartbeat for 5 minutes, stopping queue",
		zap.String("runner_id", p.RunnerID), zap.String("process_id", p.ID))

	stopTags := metrics.ToTags(tags, "reason:offline")
	a.mw.Incr("runner.process.stop", stopTags)
	if p.StartedAt != nil {
		a.mw.Timing("runner.process.latency", now.Sub(*p.StartedAt), stopTags)
	}

	if _, err := a.queueClient.EnqueueSignal(ectx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: queuesignal.NewRaw("on_inactive", map[string]any{
			"runner_id":  p.RunnerID,
			"process_id": p.ID,
			"reason":     "offline",
		}),
	}); err != nil {
		a.l.Warn("unable to enqueue on_inactive signal", zap.String("process_id", p.ID), zap.Error(err))
	}

	if err := a.helpers.StopProcessQueue(ectx, queue.ID); err != nil {
		a.l.Warn("unable to stop process queue", zap.String("process_id", p.ID), zap.Error(err))
	}
}

func (a *Activities) handleBatchProcessOffline(ectx context.Context, p *app.RunnerProcess, now time.Time, resp *BatchProcessHealthchecksResponse) (app.RunnerHealthCheck, bool) {
	if p.ProcessStatus() != app.RunnerProcessStatusOffline {
		updated, err := a.guardedProcessStatusUpdate(ectx, p, app.RunnerProcessStatusOffline, "Runner is offline and will be marked inactive in 5 minutes")
		if err != nil {
			resp.Errors++
			a.l.Warn("unable to update process status to offline", zap.String("process_id", p.ID), zap.Error(err))
			return app.RunnerHealthCheck{}, false
		}
		if !updated {
			resp.Skipped++
			return app.RunnerHealthCheck{}, false
		}
		a.l.Warn("process offline - no heartbeat for 1 minute",
			zap.String("runner_id", p.RunnerID), zap.String("process_id", p.ID))
	}
	resp.Offline++
	return a.batchHealthCheckRow(p, app.RunnerStatusError, now), true
}

func (a *Activities) handleBatchProcessActive(ectx context.Context, p *app.RunnerProcess, now time.Time, resp *BatchProcessHealthchecksResponse) (app.RunnerHealthCheck, bool) {
	if p.ProcessStatus() == app.RunnerProcessStatusOffline {
		updated, err := a.guardedProcessStatusUpdate(ectx, p, app.RunnerProcessStatusActive, "heartbeat received")
		if err != nil {
			resp.Errors++
			a.l.Warn("unable to update process status to active", zap.String("process_id", p.ID), zap.Error(err))
			return app.RunnerHealthCheck{}, false
		}
		if !updated {
			resp.Skipped++
			return app.RunnerHealthCheck{}, false
		}
		a.l.Info("process back online",
			zap.String("runner_id", p.RunnerID), zap.String("process_id", p.ID))
	}
	resp.Active++
	return a.batchHealthCheckRow(p, app.RunnerStatusActive, now), true
}

func (a *Activities) checkBatchVersionMismatch(ectx context.Context, p *app.RunnerProcess, runner app.Runner, hasRunner bool, reportedVersion string, tags map[string]string) {
	if !hasRunner || reportedVersion == "" {
		return
	}

	settings := runner.RunnerGroup.Settings
	configuredVersion := settings.ContainerImageTag
	if p.Type == app.RunnerProcessTypeMng {
		configuredVersion = settings.BinaryVersion
	}

	warning, emitLatestEvent := decideVersionWarning(configuredVersion, reportedVersion)
	if emitLatestEvent {
		a.mw.Event(&statsd.Event{
			Title: "Runner is using 'latest' version tag",
			Text: fmt.Sprintf(
				"Runner %s (org: %s) is using the 'latest' tag. configured_version=%q, reported_version=%q. Pin to a specific version to avoid drift.",
				p.RunnerID, runner.Org.Name, configuredVersion, reportedVersion,
			),
			Tags: metrics.ToTags(tags,
				metrics.ToTag("configured_version", configuredVersion),
				metrics.ToTag("heartbeat_version", reportedVersion),
			),
			SourceTypeName: "nuon-runner",
			Priority:       statsd.Normal,
			AlertType:      statsd.Error,
			AggregationKey: "runner-version-latest",
		})
	}

	current, _ := p.CompositeStatus.Metadata["version_warning"].(string)
	if current == warning {
		return
	}
	if err := generics.MergeJSONBMetadata(a.db.WithContext(ectx), &app.RunnerProcess{}, p.ID, "composite_status", map[string]any{
		"version_warning": warning,
	}); err != nil {
		a.l.Warn("unable to update version warning metadata", zap.String("process_id", p.ID), zap.Error(err))
	}
}

// guardedProcessStatusUpdate is a single-statement status transition that
// re-checks the process is still active/offline at write time, so batch checks
// can't clobber a concurrent shutdown or init transition.
func (a *Activities) guardedProcessStatusUpdate(ectx context.Context, current *app.RunnerProcess, to app.RunnerProcessStatus, desc string) (bool, error) {
	newComposite := app.NewCompositeStatus(ectx, app.Status(to))
	newComposite.StatusHumanDescription = desc
	if newComposite.Metadata == nil {
		newComposite.Metadata = map[string]any{}
	}
	for k, v := range current.CompositeStatus.Metadata {
		newComposite.Metadata[k] = v
	}
	prior := current.CompositeStatus
	prior.History = nil
	newComposite.History = append([]app.CompositeStatus{prior}, current.CompositeStatus.History...)
	if len(newComposite.History) > 25 {
		newComposite.History = newComposite.History[:25]
	}

	res := a.db.WithContext(ectx).
		Model(&app.RunnerProcess{}).
		Where("id = ?", current.ID).
		Where("composite_status->>'status' IN ?", []string{
			string(app.RunnerProcessStatusActive),
			string(app.RunnerProcessStatusOffline),
		}).
		Update("composite_status", newComposite)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (a *Activities) batchHealthCheckRow(p *app.RunnerProcess, status app.RunnerStatus, now time.Time) app.RunnerHealthCheck {
	return app.RunnerHealthCheck{
		ID:           domains.NewRunnerHealthCheckID(),
		CreatedByID:  p.CreatedByID,
		CreatedAt:    now,
		UpdatedAt:    now,
		RunnerID:     p.RunnerID,
		ProcessID:    p.ID,
		RunnerStatus: status,
	}
}
