package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"go.temporal.io/sdk/activity"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/runnerunhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const (
	defaultBatchRunnerLimit  = 200
	batchHeartbeatEvery      = 50
	orgSignalsQueueName      = "org-signals"
	runnerHealthCheckCounter = "runner.health_check"
)

type BatchRunnerHealthchecksRequest struct {
	OrgID    string `validate:"required"`
	CursorID string `json:"cursor_id"`
	Limit    int    `json:"limit"`
}

type BatchRunnerHealthchecksResponse struct {
	NextCursorID   string `json:"next_cursor_id"`
	Done           bool   `json:"done"`
	Checked        int    `json:"checked"`
	Healthy        int    `json:"healthy"`
	Unhealthy      int    `json:"unhealthy"`
	Skipped        int    `json:"skipped"`
	AlertsEnqueued int    `json:"alerts_enqueued"`
	AlertsDeduped  int    `json:"alerts_deduped"`
	Errors         int    `json:"errors"`
}

type runnerAlert struct {
	runner    app.Runner
	offlineAt time.Time
	reason    string
	tags      map[string]string
}

// BatchRunnerHealthchecks checks one keyset page of an org's runners inline,
// replacing the per-runner runner_healthcheck signal fan-out.
//
// @temporal-gen-v2 activity
// @by-field OrgID
// @start-to-close-timeout 5m
// @heartbeat-timeout 60s
func (a *Activities) BatchRunnerHealthchecks(ctx context.Context, req BatchRunnerHealthchecksRequest) (*BatchRunnerHealthchecksResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = defaultBatchRunnerLimit
	}
	resp := &BatchRunnerHealthchecksResponse{}

	var runners []app.Runner
	if res := a.db.WithContext(ctx).
		Select("id", "display_name", "org_id", "created_by_id", "status", "status_v2", "runner_group_id").
		Preload("Org").
		Preload("RunnerGroup").
		Where("org_id = ?", req.OrgID).
		Where("id > ?", req.CursorID).
		Order("id").
		Limit(limit).
		Find(&runners); res.Error != nil {
		return nil, fmt.Errorf("unable to list runners for org %s: %w", req.OrgID, res.Error)
	}

	resp.Done = len(runners) < limit
	if len(runners) == 0 {
		return resp, nil
	}
	resp.NextCursorID = runners[len(runners)-1].ID

	runnerIDs := make([]string, 0, len(runners))
	for _, r := range runners {
		runnerIDs = append(runnerIDs, r.ID)
	}
	presence, err := a.activeProcessPresence(ctx, runnerIDs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var alerts []runnerAlert

	for i := range runners {
		if i%batchHeartbeatEvery == 0 {
			activity.RecordHeartbeat(ctx, runners[i].ID)
		}
		r := runners[i]
		resp.Checked++

		d := decideRunnerHealth(now, &r, presence[r.ID])
		tags := runnerHealthTags(&r, presence[r.ID], d)

		switch d.Result {
		case "skipped":
			resp.Skipped++
			a.mw.Incr(runnerHealthCheckCounter, metrics.ToTags(tags, metrics.ToTag("result", "skipped")))
			continue
		case "healthy":
			resp.Healthy++
		case "unhealthy":
			resp.Unhealthy++
		default:
			continue
		}

		ectx := cctx.SetOrgIDContext(ctx, r.OrgID)
		ectx = cctx.SetAccountIDContext(ectx, r.CreatedByID)

		if err := a.applyRunnerHealthDecision(ectx, &r, d, now); err != nil {
			resp.Errors++
			a.l.Warn("unable to apply runner healthcheck decision",
				zap.String("runner_id", r.ID), zap.Error(err))
			continue
		}

		if d.Alert {
			alerts = append(alerts, runnerAlert{runner: r, offlineAt: d.AlertOfflineAt, reason: d.Reason, tags: tags})
		}

		a.mw.Incr(runnerHealthCheckCounter, metrics.ToTags(tags, metrics.ToTag("result", d.Result)))
	}

	if len(alerts) > 0 {
		a.emitRunnerAlerts(ctx, req.OrgID, alerts, resp)
	}

	return resp, nil
}

func (a *Activities) activeProcessPresence(ctx context.Context, runnerIDs []string) (map[string]runnerProcessPresence, error) {
	var rows []struct {
		RunnerID string
		Type     app.RunnerProcessType
		Status   string
	}
	if res := a.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (runner_id, type) runner_id, type, composite_status->>'status' AS status
		FROM runner_processes
		WHERE runner_id IN ? AND type IN ? AND deleted_at = 0
		ORDER BY runner_id, type, created_at DESC`,
		runnerIDs,
		[]string{
			string(app.RunnerProcessTypeBuild),
			string(app.RunnerProcessTypeInstall),
			string(app.RunnerProcessTypeMng),
		}).Scan(&rows); res.Error != nil {
		return nil, fmt.Errorf("unable to resolve current runner processes: %w", res.Error)
	}

	presence := make(map[string]runnerProcessPresence, len(runnerIDs))
	for _, id := range runnerIDs {
		presence[id] = runnerProcessPresence{MngChecked: true}
	}
	for _, row := range rows {
		p := presence[row.RunnerID]
		active := row.Status == string(app.RunnerProcessStatusActive)
		switch row.Type {
		case app.RunnerProcessTypeBuild:
			p.HasActiveBuild = active
		case app.RunnerProcessTypeInstall:
			p.HasActiveInstall = active
		case app.RunnerProcessTypeMng:
			p.HasActiveMng = active
		}
		presence[row.RunnerID] = p
	}
	return presence, nil
}

// applyRunnerHealthDecision performs the decision's writes in the signal's
// order: mng metadata, offline_ts arm/clear, legacy status (fail-fast), then
// status v2.
func (a *Activities) applyRunnerHealthDecision(ectx context.Context, r *app.Runner, d runnerHealthDecision, now time.Time) error {
	if d.SetMissingMng != nil {
		if err := a.statusActivities.UpdateRunnerStatusV2Metadata(ectx, statusactivities.UpdateRunnerStatusV2MetadataRequest{
			RunnerID: r.ID,
			Metadata: map[string]any{"missing_mng_process": *d.SetMissingMng},
		}); err != nil {
			return fmt.Errorf("unable to update management process status metadata: %w", err)
		}
	}

	if d.SetOfflineTS {
		if err := a.statusActivities.UpdateRunnerStatusV2Metadata(ectx, statusactivities.UpdateRunnerStatusV2MetadataRequest{
			RunnerID: r.ID,
			Metadata: map[string]any{
				app.RunnerOfflineTSMetadataKey:         now.Unix(),
				app.RunnerOfflineFromStatusMetadataKey: d.OfflineFromStatus,
			},
		}); err != nil {
			return fmt.Errorf("unable to set runner offline metadata: %w", err)
		}
	}
	if d.ClearOfflineTS {
		if err := a.statusActivities.UpdateRunnerStatusV2Metadata(ectx, statusactivities.UpdateRunnerStatusV2MetadataRequest{
			RunnerID: r.ID,
			Metadata: map[string]any{
				app.RunnerOfflineTSMetadataKey:         nil,
				app.RunnerOfflineFromStatusMetadataKey: nil,
			},
		}); err != nil {
			return fmt.Errorf("unable to clear runner offline metadata: %w", err)
		}
	}

	if d.UpdateLegacy {
		// Guarded write: the decision was computed from a read that may predate
		// a reconcile marking this runner disabled. Overwriting that would pin
		// an intentionally-disabled runner to offline, and since the skip
		// conditions match on status it would never recover.
		res := a.db.WithContext(ectx).
			Model(&app.Runner{ID: r.ID}).
			Where("status <> ?", app.RunnerStatusDisabled).
			Updates(app.Runner{
				Status:            d.TargetStatus,
				StatusDescription: d.Reason,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update runner status: %w", res.Error)
		}
		if res.RowsAffected < 1 {
			// Runner went disabled under us; leave its status v2 alone too so
			// the two columns cannot disagree.
			return nil
		}
	}
	if d.UpdateV2 {
		if err := a.statusActivities.UpdateRunnerStatusV2(ectx, statusactivities.UpdateRunnerStatusV2Request{
			RunnerID:          r.ID,
			Status:            d.TargetStatus,
			StatusDescription: d.Reason,
		}); err != nil {
			return fmt.Errorf("unable to update runner status v2: %w", err)
		}
	}

	return nil
}

func (a *Activities) emitRunnerAlerts(ctx context.Context, orgID string, alerts []runnerAlert, resp *BatchRunnerHealthchecksResponse) {
	installIDs := make([]string, 0)
	for _, al := range alerts {
		if al.runner.RunnerGroup.OwnerType == "installs" {
			installIDs = append(installIDs, al.runner.RunnerGroup.OwnerID)
		}
	}
	installsByID := make(map[string]app.Install)
	if len(installIDs) > 0 {
		var installs []app.Install
		if res := a.db.WithContext(ctx).
			Select("id", "name", "app_id", "created_by_id").
			Preload("App").
			Preload("CreatedBy").
			Where("id IN ?", installIDs).
			Find(&installs); res.Error != nil {
			a.l.Warn("unable to load installs for runner unhealthy alerts", zap.Error(res.Error))
		}
		for _, in := range installs {
			installsByID[in.ID] = in
		}
	}

	queue, err := a.queueClient.GetQueueByOwnerAndName(ctx, orgID, "orgs", orgSignalsQueueName)
	if err != nil {
		resp.Errors += len(alerts)
		a.l.Warn("unable to resolve org-signals queue for runner unhealthy alerts",
			zap.String("org_id", orgID), zap.Error(err))
		return
	}

	for _, al := range alerts {
		r := al.runner
		var install *app.Install
		if in, ok := installsByID[r.RunnerGroup.OwnerID]; ok {
			install = &in
		}
		fromStatus := runnerOfflineFromStatus(&r)
		event, ownerName := runnerOfflineEvent(&r, install, fromStatus, al.reason, al.tags)

		ectx := cctx.SetOrgIDContext(ctx, r.OrgID)
		ectx = cctx.SetAccountIDContext(ectx, r.CreatedByID)
		enqResp, err := a.queueClient.EnqueueSignal(ectx, &queueclient.EnqueueSignalRequest{
			QueueID:        queue.ID,
			OwnerID:        r.ID,
			OwnerType:      "runners",
			IdempotencyKey: fmt.Sprintf("runner-unhealthy:%s:%d", r.ID, al.offlineAt.Unix()),
			Signal: &runnerunhealthy.Signal{
				RunnerID:             r.ID,
				RunnerName:           r.DisplayName,
				OrgID:                r.OrgID,
				OrgName:              r.Org.Name,
				FromStatus:           fromStatus,
				ToStatus:             app.RunnerStatusOffline,
				Reason:               al.reason,
				RunnerGroupID:        r.RunnerGroupID,
				RunnerGroupType:      r.RunnerGroup.Type,
				RunnerGroupOwnerID:   r.RunnerGroup.OwnerID,
				RunnerGroupOwnerType: r.RunnerGroup.OwnerType,
				RunnerGroupOwnerName: ownerName,
			},
		})
		if err != nil {
			resp.Errors++
			a.l.Warn("unable to enqueue runner unhealthy signal",
				zap.String("runner_id", r.ID), zap.Error(err))
			continue
		}
		if enqResp.Deduplicated {
			resp.AlertsDeduped++
			continue
		}
		resp.AlertsEnqueued++
		a.mw.Event(event)
	}
}

func runnerHealthTags(r *app.Runner, presence runnerProcessPresence, d runnerHealthDecision) map[string]string {
	tags := map[string]string{
		"runner_id":     r.ID,
		"runner_type":   string(r.RunnerGroup.Type),
		"runner_status": string(r.Status),
		"org_id":        r.OrgID,
		"org_name":      r.Org.Name,
	}
	if r.RunnerGroup.OwnerType == "installs" {
		tags["install_id"] = r.RunnerGroup.OwnerID
	}
	if d.Result == "skipped" {
		return tags
	}

	switch r.RunnerGroup.Type {
	case app.RunnerGroupTypeOrg:
		tags["missing_build_process"] = fmt.Sprintf("%t", !presence.HasActiveBuild)
	case app.RunnerGroupTypeInstall:
		tags["missing_install_process"] = fmt.Sprintf("%t", !presence.HasActiveInstall)
		tags["missing_mng_process"] = fmt.Sprintf("%t", presence.MngChecked && !presence.HasActiveMng)
	}
	return tags
}

func runnerOfflineFromStatus(r *app.Runner) app.RunnerStatus {
	if raw, ok := r.StatusV2.Metadata[app.RunnerOfflineFromStatusMetadataKey].(string); ok && raw != "" {
		return app.RunnerStatus(raw)
	}
	return app.RunnerStatusUnknown
}

func runnerOfflineEvent(r *app.Runner, install *app.Install, fromStatus app.RunnerStatus, reason string, tags map[string]string) (*statsd.Event, string) {
	eventTags := []string{
		metrics.ToTag("runner_id", r.ID),
		metrics.ToTag("runner_type", string(r.RunnerGroup.Type)),
		metrics.ToTag("org_id", r.OrgID),
		metrics.ToTag("org_name", r.Org.Name),
	}

	title := fmt.Sprintf("Runner went offline (type: %s)", string(r.RunnerGroup.Type))
	text := fmt.Sprintf(
		"Runner %s (org: %s) transitioned from %s to offline.\nReason: %s",
		r.ID, r.Org.Name, fromStatus, reason,
	)
	ownerName := ""
	if r.RunnerGroup.OwnerType == "orgs" {
		ownerName = r.Org.Name
	}
	if install != nil {
		ownerName = install.Name
		eventTags = append(eventTags,
			metrics.ToTag("install_id", install.ID),
			metrics.ToTag("install_name", install.Name),
			metrics.ToTag("app_id", install.AppID),
			metrics.ToTag("app_name", install.App.Name),
			metrics.ToTag("created_by", install.CreatedBy.Email),
		)
		text = fmt.Sprintf(
			"Runner %s (org: %s, app: %s, install: %s) transitioned from %s to offline.\nReason: %s\nInstall created by: %s",
			r.ID, r.Org.Name, install.App.Name, install.Name, fromStatus, reason, install.CreatedBy.Email,
		)
	}

	return &statsd.Event{
		Title:          title,
		Text:           text,
		Tags:           eventTags,
		SourceTypeName: "nuon-runner",
		Priority:       statsd.Normal,
		AlertType:      statsd.Error,
		AggregationKey: "runner-health-check",
	}, ownerName
}
