package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/supportsnapshot"
	"go.uber.org/zap"
)

const (
	maxSnapshotLogsPerJob = 2000
	maxSnapshotLogsTotal  = 20000
)

var snapshotLogFields = map[string]bool{
	"runner_job.id": true, "runner_job.type": true, "runner_job.group": true,
	"runner_job.operation": true, "component": true, "step": true,
	"attempt": true, "exit_code": true,
}

func (p *portalServer) supportSnapshot(w http.ResponseWriter, r *http.Request) {
	includeState := false
	if value := r.URL.Query().Get("include_state"); value != "" {
		var err error
		includeState, err = strconv.ParseBool(value)
		if err != nil {
			writeAPIError(w, fmt.Errorf("include_state must be true or false"), http.StatusBadRequest)
			return
		}
	}
	snapshot, err := p.buildSupportSnapshot(r, includeState)
	if err != nil {
		writeAPIError(w, err, registrationErrorStatus(err))
		return
	}
	filename := fmt.Sprintf("nuon-support-%s-%s.tar.zst", safeFilenamePart(snapshot.Registration.DeploymentID), snapshot.CapturedAt.Format("20060102T150405Z"))
	w.Header().Set("Content-Type", supportsnapshot.ArchiveContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if _, err := supportsnapshot.Write(w, snapshot, supportsnapshot.Producer{Name: "bundle-portal", RunnerVersion: runnerVersion(snapshot.Runner)}); err != nil {
		p.logger.Error("write support snapshot", zap.Error(err))
	}
}

func (p *portalServer) buildSupportSnapshot(r *http.Request, includeState bool) (supportsnapshot.Snapshot, error) {
	registration, err := p.buildInstallationRegistration(r.Context())
	if err != nil {
		return supportsnapshot.Snapshot{}, err
	}
	now := time.Now().UTC()
	redaction := "support-snapshot-safe-fields-v1"
	if includeState {
		redaction += "+customer-selected-raw-state-unredacted"
	}
	report := supportsnapshot.CollectionReport{SchemaVersion: supportsnapshot.SchemaVersion, Redaction: redaction, Unavailable: map[string]string{}, Truncated: map[string]int64{}}
	snapshot := supportsnapshot.Snapshot{SchemaVersion: supportsnapshot.SchemaVersion, CapturedAt: now, Registration: registration, IncludeState: includeState, Collection: report}

	if includeState {
		p.collectState(r.Context(), &snapshot)
	}
	collectHeartbeat(p, r.Context(), &snapshot)
	collectJSON(p, r.Context(), "catalog", operation.CatalogKey, &snapshot.Catalog, &snapshot.Collection)
	collectJSON(p, r.Context(), "active_bundle", operation.BundleKey, &snapshot.ActiveBundle, &snapshot.Collection)
	p.collectStagedBundle(r.Context(), &snapshot)
	p.collectBundleHistory(r.Context(), &snapshot)
	collectJSON(p, r.Context(), "health", "health/latest.json", &snapshot.Health, &snapshot.Collection)
	collectJSON(p, r.Context(), "health_transitions", "health/transitions.json", &snapshot.HealthTransitions, &snapshot.Collection)
	collectJSON(p, r.Context(), "current_inputs", customermanaged.CapturedInputsKey, &snapshot.CurrentInputs, &snapshot.Collection)
	collectJSON(p, r.Context(), "roles", customermanaged.CapturedRolesKey, &snapshot.Roles, &snapshot.Collection)
	p.collectRuns(r, &snapshot)
	p.collectSnapshotLogs(r, &snapshot)
	if len(snapshot.Collection.Unavailable) == 0 {
		snapshot.Collection.Unavailable = nil
	}
	if len(snapshot.Collection.Truncated) == 0 {
		snapshot.Collection.Truncated = nil
	}
	return snapshot, nil
}

func (p *portalServer) collectStagedBundle(ctx context.Context, snapshot *supportsnapshot.Snapshot) {
	candidate, _, found, err := p.latestBundleCandidateRecord(ctx)
	if err != nil {
		snapshot.Collection.Unavailable["staged_bundle"] = err.Error()
		return
	}
	if !found {
		snapshot.Collection.Unavailable["staged_bundle"] = "not found"
		return
	}
	snapshot.StagedBundle = &candidate
	snapshot.Collection.Included = append(snapshot.Collection.Included, "staged_bundle")
}

func (p *portalServer) collectState(ctx context.Context, snapshot *supportsnapshot.Snapshot) {
	state := &supportsnapshot.CapturedState{}
	for key, target := range map[string]*json.RawMessage{
		"status.json": &state.Status,
		"report.json": &state.Report,
	} {
		raw, ok, err := p.store.Get(ctx, key)
		if err != nil {
			snapshot.Collection.Unavailable["state."+strings.TrimSuffix(key, ".json")] = err.Error()
			continue
		}
		if !ok {
			snapshot.Collection.Unavailable["state."+strings.TrimSuffix(key, ".json")] = "not found"
			continue
		}
		if !json.Valid(raw) {
			snapshot.Collection.Unavailable["state."+strings.TrimSuffix(key, ".json")] = "invalid JSON"
			continue
		}
		*target = append((*target)[:0], raw...)
	}
	if len(state.Status) == 0 && len(state.Report) == 0 {
		snapshot.Collection.Unavailable["state"] = "not found"
		return
	}
	snapshot.State = state
	snapshot.Collection.Included = append(snapshot.Collection.Included, "state")
}

func collectJSON[T any](p *portalServer, ctx context.Context, section, key string, target *T, report *supportsnapshot.CollectionReport) {
	raw, ok, err := p.store.Get(ctx, key)
	if err != nil {
		report.Unavailable[section] = err.Error()
		return
	}
	if !ok {
		report.Unavailable[section] = "not found"
		return
	}
	if err := json.Unmarshal(raw, target); err != nil {
		report.Unavailable[section] = err.Error()
		return
	}
	report.Included = append(report.Included, section)
}

func collectHeartbeat(p *portalServer, ctx context.Context, snapshot *supportsnapshot.Snapshot) {
	raw, ok, err := p.store.Get(ctx, customermanaged.RunnerHeartbeatKey)
	if err == nil && !ok {
		raw, ok, err = p.store.Get(ctx, customermanaged.LegacyRunnerHeartbeatKey)
	}
	if err != nil {
		snapshot.Collection.Unavailable["runner"] = err.Error()
		return
	}
	if !ok {
		snapshot.Collection.Unavailable["runner"] = "not found"
		return
	}
	if err := json.Unmarshal(raw, &snapshot.Runner); err != nil {
		snapshot.Collection.Unavailable["runner"] = err.Error()
		return
	}
	snapshot.Collection.Included = append(snapshot.Collection.Included, "runner")
}

func (p *portalServer) collectBundleHistory(ctx context.Context, snapshot *supportsnapshot.Snapshot) {
	keys, err := p.store.List(ctx, operation.BundlesPrefix)
	if err != nil {
		snapshot.Collection.Unavailable["bundle_history"] = err.Error()
		return
	}
	for _, key := range keys {
		if !strings.HasSuffix(key, ".json") {
			continue
		}
		var info operation.BundleInfo
		raw, ok, err := p.store.Get(ctx, key)
		if err != nil || !ok {
			continue
		}
		if err := json.Unmarshal(raw, &info); err != nil {
			snapshot.Collection.Unavailable["bundle_history"] = err.Error()
			return
		}
		snapshot.BundleHistory = append(snapshot.BundleHistory, info)
	}
	sort.Slice(snapshot.BundleHistory, func(i, j int) bool {
		return snapshot.BundleHistory[i].ActivatedAt.After(snapshot.BundleHistory[j].ActivatedAt)
	})
	snapshot.Collection.Included = append(snapshot.Collection.Included, "bundle_history")
}

func (p *portalServer) collectRuns(r *http.Request, snapshot *supportsnapshot.Snapshot) {
	runs, err := p.listRuns(r)
	if err != nil {
		snapshot.Collection.Unavailable["runs"] = err.Error()
		return
	}
	for _, run := range runs {
		complete, err := p.readRun(r, run.RunID)
		if err != nil {
			snapshot.Collection.Unavailable["run_events/"+run.RunID] = err.Error()
		} else if complete != nil {
			run = *complete
		}
		captured := snapshotRun(run)
		for i := range captured.Steps {
			step := &captured.Steps[i]
			resultID := step.JobID
			if resultID == "" {
				resultID = run.RunID + "--" + step.ID
			}
			result, ok, resultErr := p.readStepResult(r.Context(), resultID)
			if resultErr != nil {
				snapshot.Collection.Unavailable["plans/"+resultID] = resultErr.Error()
				continue
			}
			if ok && result.Kind != "unknown" {
				step.Plan = &supportsnapshot.StepPlan{Kind: result.Kind, Content: result.Content}
			}
		}
		snapshot.Runs = append(snapshot.Runs, captured)
	}
	snapshot.Collection.Included = append(snapshot.Collection.Included, "runs")
}

func snapshotRun(run portalRun) supportsnapshot.Run {
	result := supportsnapshot.Run{RunID: run.RunID, DispatchID: run.DispatchID, RefID: run.RefID, RefKind: run.RefKind, RefName: run.RefName, Source: run.Source, Status: run.Status, Error: run.Error, BundleDigest: run.BundleDigest, PreviousRunID: run.PreviousRunID, StartedAt: run.StartedAt, FinishedAt: run.FinishedAt, ResultDirective: run.ResultDirective}
	for _, step := range run.Steps {
		result.Steps = append(result.Steps, supportsnapshot.RunStep{ID: step.ID, Name: step.Name, Kind: step.Kind, JobID: step.JobID, Status: step.Status, Error: step.Error, StartedAt: step.StartedAt, FinishedAt: step.FinishedAt, SourceRunID: step.SourceRunID, ResultDirective: step.ResultDirective, Description: step.Description, Drift: step.Drift})
	}
	return result
}

func (p *portalServer) collectSnapshotLogs(r *http.Request, snapshot *supportsnapshot.Snapshot) {
	summaries := p.jobSummaries(r)
	jobs := make([]jobLogSummary, 0, len(summaries))
	for _, summary := range summaries {
		jobs = append(jobs, summary)
	}
	sort.SliceStable(jobs, func(i, j int) bool {
		if jobs[i].StartedAt == nil {
			return false
		}
		if jobs[j].StartedAt == nil {
			return true
		}
		return jobs[i].StartedAt.After(*jobs[j].StartedAt)
	})
	remaining := maxSnapshotLogsTotal
	for _, job := range jobs {
		entries, ok, err := p.readJobLogEntries(r.Context(), job.JobID)
		if err != nil {
			snapshot.Collection.Unavailable["logs/"+job.JobID] = err.Error()
			continue
		}
		if !ok {
			continue
		}
		keep := len(entries)
		if keep > maxSnapshotLogsPerJob {
			keep = maxSnapshotLogsPerJob
		}
		if keep > remaining {
			keep = remaining
		}
		dropped := len(entries) - keep
		out := supportsnapshot.JobLog{JobID: job.JobID, RunID: job.RunID, Name: job.Name, Status: job.Status, StartedAt: job.StartedAt, Total: len(entries), Truncated: dropped > 0}
		for _, entry := range entries[len(entries)-keep:] {
			if entry.Raw != "" {
				continue
			}
			fields := map[string]any{}
			for key, value := range entry.Fields {
				if snapshotLogFields[key] {
					fields[key] = value
				}
			}
			if len(fields) == 0 {
				fields = nil
			}
			out.Entries = append(out.Entries, supportsnapshot.LogEntry{Time: entry.Time, Level: entry.Level, Msg: entry.Msg, Fields: fields})
		}
		if skippedMalformed := keep - len(out.Entries); skippedMalformed > 0 {
			dropped += skippedMalformed
			out.Truncated = true
		}
		if dropped > 0 {
			snapshot.Collection.Truncated["logs/"+job.JobID] = int64(dropped)
		}
		snapshot.Logs = append(snapshot.Logs, out)
		remaining -= keep
	}
	snapshot.Collection.Included = append(snapshot.Collection.Included, "logs")
}

func runnerVersion(heartbeat *customermanaged.RunnerHeartbeat) string {
	if heartbeat == nil {
		return ""
	}
	return heartbeat.Version
}
func safeFilenamePart(value string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, value)
}
