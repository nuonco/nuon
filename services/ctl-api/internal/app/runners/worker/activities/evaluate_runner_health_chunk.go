package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// heartBeatTimeout mirrors the runner healthcheck signal's threshold: a heartbeat
// older than this (or missing) marks the runner as errored.
const heartBeatTimeout = 15 * time.Second

type EvaluateRunnerHealthChunkRequest struct {
	RunnerIDs []string `validate:"required"`
}

type RunnerStatusChange struct {
	RunnerID  string           `json:"runner_id"`
	OldStatus app.RunnerStatus `json:"old_status"`
	NewStatus app.RunnerStatus `json:"new_status"`
}

type EvaluateRunnerHealthChunkResponse struct {
	Changes []RunnerStatusChange `json:"changes"`
}

// EvaluateRunnerHealthChunk evaluates a chunk of runners against their latest
// heartbeat, writes a per-tick health-check record for each (preserving the
// signal's behavior), updates warnings, and returns only the runners whose status
// changed (status writes are applied by the sweep workflow via the existing
// status activities). All reads/records are batched server-side so the sweep
// workflow's history stays small regardless of runner count.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) EvaluateRunnerHealthChunk(ctx context.Context, req EvaluateRunnerHealthChunkRequest) (*EvaluateRunnerHealthChunkResponse, error) {
	if len(req.RunnerIDs) == 0 {
		return &EvaluateRunnerHealthChunkResponse{}, nil
	}

	var runners []app.Runner
	if res := a.db.WithContext(ctx).
		Preload("RunnerGroup").
		Preload("RunnerGroup.Settings").
		Where("id IN ?", req.RunnerIDs).
		Find(&runners); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to load runners")
	}

	heartbeats, err := a.latestHeartbeatsByRunner(ctx, req.RunnerIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load heartbeats")
	}

	now := time.Now()
	minHeartBeatTS := now.Add(-heartBeatTimeout)

	checks := make([]app.RunnerHealthCheck, 0, len(runners))
	resp := &EvaluateRunnerHealthChunkResponse{}

	for i := range runners {
		runner := &runners[i]

		process := app.HeartBeatProcessForRunnerGroupType(runner.RunnerGroup.Type)
		hb := heartbeats[runner.ID][process]
		if hb == nil {
			// Fallback to the unknown process type, matching getMostRecentHeartBeat.
			hb = heartbeats[runner.ID][app.RunnerProcessTypeUnknown]
		}

		newStatus := app.RunnerStatusActive
		if hb == nil || hb.CreatedAt.Before(minHeartBeatTS) {
			newStatus = app.RunnerStatusError
		}

		checks = append(checks, app.RunnerHealthCheck{
			RunnerID:     runner.ID,
			RunnerStatus: newStatus,
		})

		// Warnings are recomputed every tick (same as the signal).
		warnings, isAliasTag := computeRunnerWarnings(runner, hb)
		if err := a.UpdateWarnings(ctx, UpdateWarningsRequest{
			RunnerID:   runner.ID,
			Warnings:   warnings,
			IsAliasTag: isAliasTag,
		}); err != nil {
			return nil, errors.Wrapf(err, "unable to update warnings for runner %s", runner.ID)
		}

		if runner.Status != newStatus {
			resp.Changes = append(resp.Changes, RunnerStatusChange{
				RunnerID:  runner.ID,
				OldStatus: runner.Status,
				NewStatus: newStatus,
			})
		}
	}

	if len(checks) > 0 {
		if res := a.chDB.WithContext(ctx).Create(&checks); res.Error != nil {
			return nil, errors.Wrap(res.Error, "unable to create health check records")
		}
	}

	return resp, nil
}

// latestHeartbeatsByRunner returns the most recent heartbeat per (runner, process)
// for the given runners in a single ClickHouse query.
func (a *Activities) latestHeartbeatsByRunner(ctx context.Context, runnerIDs []string) (map[string]map[app.RunnerProcessType]*app.RunnerHeartBeat, error) {
	type hbRow struct {
		RunnerID  string
		Process   app.RunnerProcessType
		Version   string
		CreatedAt time.Time `gorm:"column:latest_created_at"`
	}

	// NOTE: the max() alias must NOT be "created_at" — that collides with the
	// column referenced inside argMax(version, created_at), causing ClickHouse to
	// resolve it to the max() aggregate and error with "aggregate function found
	// inside another aggregate function" (code 184).
	var rows []hbRow
	if res := a.chDB.WithContext(ctx).
		Model(&app.RunnerHeartBeat{}).
		Select("runner_id, process, argMax(version, created_at) as version, max(created_at) as latest_created_at").
		Where("runner_id IN ?", runnerIDs).
		Group("runner_id, process").
		Scan(&rows); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to query heartbeats")
	}

	out := make(map[string]map[app.RunnerProcessType]*app.RunnerHeartBeat, len(rows))
	for i := range rows {
		r := rows[i]
		if out[r.RunnerID] == nil {
			out[r.RunnerID] = map[app.RunnerProcessType]*app.RunnerHeartBeat{}
		}
		out[r.RunnerID][r.Process] = &app.RunnerHeartBeat{
			RunnerID:  r.RunnerID,
			Process:   r.Process,
			Version:   r.Version,
			CreatedAt: r.CreatedAt,
		}
	}

	return out, nil
}

// computeRunnerWarnings mirrors the runner healthcheck signal's computeWarnings.
func computeRunnerWarnings(runner *app.Runner, heartbeat *app.RunnerHeartBeat) (pq.StringArray, bool) {
	var warnings pq.StringArray
	if heartbeat == nil {
		return warnings, false
	}

	expectedVersion := runner.RunnerGroup.Settings.ContainerImageTag
	reportedVersion := heartbeat.Version

	// A non-semver configured tag (e.g. "latest") is an alias; the runner reports
	// the resolved version, so skip the mismatch warning.
	if expectedVersion != "" {
		if _, err := semver.NewVersion(expectedVersion); err != nil {
			return warnings, true
		}
	}

	if expectedVersion != "" && reportedVersion != "" && expectedVersion != reportedVersion {
		warnings = append(warnings, fmt.Sprintf("Reported version (%s) does not match configured version (%s).", reportedVersion, expectedVersion))
	}

	return warnings, false
}
