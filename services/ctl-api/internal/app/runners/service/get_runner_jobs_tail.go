package service

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// Long-poll tuning for job pickup. Pickup is event-driven: a Postgres trigger
// (migration 112) fires NOTIFY on the queued→available transition, which the
// pod-level RunnerJobNotifyListener fans out to parked handlers via the wake
// registry, so the common case returns in ~ms. jobTailBackstopInterval is the
// sparse safety poll that bounds worst-case latency when a notify is dropped
// (listener reconnect, pod restart, RDS failover) — it is NOT the primary
// mechanism, so it's deliberately slow (and cheaper than the old 1s loop).
const (
	jobTailMaxWait             = 25 * time.Second
	jobTailBackstopInterval    = 5 * time.Second
	jobTailMaxConcurrentProbes = 50
	jobTailProbeQueryTimeout   = 2 * time.Second
)

var jobTailProbeSem = make(chan struct{}, jobTailMaxConcurrentProbes)

const (
	metricJobTailHotProbeMs = "runner_job_tail.hot_probe_ms"
	metricJobTailOutcome    = "runner_job_tail.outcome"
	metricJobTailNotifyWake = "runner_job_tail.notify_wake_probe"
)

const (
	jobTailOutcomeHotHit       = "hot_hit"
	jobTailOutcomeIdleThenHit  = "idle_then_hit"
	jobTailOutcomeTimeoutEmpty = "timeout_empty"
	jobTailOutcomeClientCancel = "client_cancel"
	jobTailOutcomeError        = "error"
)

// @ID						TailRunnerJobs
// @Summary				long-poll for an available runner job
// @Description			Returns the next available job for this runner, holding the request up to ~25s when the queue is empty. Behind the `runner-job-long-poll` org feature flag; returns 404 when the flag is off so the runner can fall back to legacy 5s polling.
// @Param					runner_id	path	string	true	"runner ID"
// @Param					group		query	string	false	"job group"	Default(any)
// @Param					wait		query	string	false	"max wait for an available job (Go duration, capped server-side at 25s)"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.RunnerJob
// @Router					/v1/runners/{runner_id}/jobs/tail [get]
func (s *service) TailRunnerJobs(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get runner"))
		return
	}

	// Feature gate. 404 lets the runner SDK treat the endpoint as if it
	// doesn't exist for this org and stay on the legacy poll. 403/501
	// would force callers to surface this as a real error.
	if !runner.Org.Features[string(app.OrgFeatureRunnerJobLongPoll)] {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	groupStr := ctx.DefaultQuery("group", string(app.RunnerJobGroupAny))
	grp := app.RunnerJobGroup(groupStr)

	wait := jobTailMaxWait
	if w := ctx.Query("wait"); w != "" {
		d, perr := time.ParseDuration(w)
		if perr != nil {
			ctx.Error(stderr.NewInvalidRequest(errors.Wrap(perr, "invalid `wait` duration")))
			return
		}
		if d > 0 && d < wait {
			wait = d
		}
	}

	startedAt := time.Now()
	deadline := startedAt.Add(wait)
	firstIter := true
	// Subscribe BEFORE the first probe so a NOTIFY that fires between our probe
	// and entering the select isn't missed — the buffered wake channel holds it
	// and the next select drains it immediately.
	wakeCh, unsubscribe := s.runnerJobWake.Subscribe(runnerID)
	defer unsubscribe()

	for {
		probeStart := time.Now()
		job, qerr := s.tailJobProbe(ctx.Request.Context(), runnerID, grp)
		if qerr != nil {
			s.emitJobTailExit(jobTailOutcomeError)
			ctx.Error(errors.Wrap(qerr, "unable to probe runner job tail"))
			return
		}

		if job != nil {
			outcome := jobTailOutcomeIdleThenHit
			if firstIter {
				s.mw.Timing(metricJobTailHotProbeMs, time.Since(probeStart), nil)
				outcome = jobTailOutcomeHotHit
			}
			s.emitJobTailExit(outcome)
			s.emitRunnerJobPickupAge([]*app.RunnerJob{job}, pickupPathLongPoll)
			ctx.JSON(http.StatusOK, []*app.RunnerJob{job})
			return
		}

		firstIter = false

		remaining := time.Until(deadline)
		if remaining <= 0 {
			s.emitJobTailExit(jobTailOutcomeTimeoutEmpty)
			ctx.JSON(http.StatusOK, []*app.RunnerJob{})
			return
		}

		// Wait for whichever comes first: a NOTIFY wake (re-probe in ~ms), the
		// sparse backstop tick, or the client/deadline going away. No exponential
		// backoff — the notify carries the hot path.
		sleep := jobTailBackstopInterval + jitter(jobTailBackstopInterval)
		if sleep > remaining {
			sleep = remaining
		}
		select {
		case <-ctx.Request.Context().Done():
			s.emitJobTailExit(jobTailOutcomeClientCancel)
			return
		case <-wakeCh:
			s.mw.Count(metricJobTailNotifyWake, 1, nil)
		case <-time.After(sleep):
		}
	}
}

func (s *service) emitJobTailExit(result string) {
	s.mw.Count(metricJobTailOutcome, 1, []string{"result:" + result})
}

// tailJobProbe runs a single bounded Postgres query for the next available
// job. Semaphore protects the DB pool from a burst of long-pollers all
// firing probes at the same instant.
func (s *service) tailJobProbe(parent context.Context, runnerID string, grp app.RunnerJobGroup) (*app.RunnerJob, error) {
	select {
	case jobTailProbeSem <- struct{}{}:
	case <-parent.Done():
		return nil, parent.Err()
	}
	defer func() { <-jobTailProbeSem }()

	ctx, cancel := context.WithTimeout(parent, jobTailProbeQueryTimeout)
	defer cancel()

	where := app.RunnerJob{
		RunnerID: runnerID,
		Status:   app.RunnerJobStatusAvailable,
	}
	if grp != app.RunnerJobGroupAny {
		where.Group = grp
	}

	var job app.RunnerJob
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithDisableViews).
		Where(where).
		Order("created_at desc").
		Limit(1).
		Take(&job)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(res.Error, "unable to query runner job tail")
	}
	return &job, nil
}
