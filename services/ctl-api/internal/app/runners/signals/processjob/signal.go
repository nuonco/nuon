package processjob

import (
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	pkgerrors "github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "process_job"

const (
	minPollPeriod = time.Second
	maxPollPeriod = 10 * time.Second
)

// pollPeriod backs off 1s→10s (1,2,4,8,10,...) so fast jobs are detected quickly
// while long jobs stay cheap on workflow history.
func pollPeriod(attempt int) time.Duration {
	if attempt < 0 || attempt > 4 {
		return maxPollPeriod
	}
	d := minPollPeriod << attempt
	if d > maxPollPeriod {
		return maxPollPeriod
	}
	return d
}

type Signal struct {
	RunnerID string `json:"runner_id"`
	JobID    string `json:"job_id"`

	cfg *internal.Config
	mw  tmetrics.Writer
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithParams = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) WithParams(params *signal.Params) {
	s.cfg = params.Cfg
	s.mw, _ = tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MW),
		tmetrics.WithTags(map[string]string{
			"namespace": "runners",
			"context":   "signal",
		}))
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}
	if s.JobID == "" {
		return errors.New("job_id is required")
	}

	// Validate runner exists
	_, err := activities.AwaitGet(ctx, activities.GetRequest{RunnerID: s.RunnerID})
	if err != nil {
		_ = activities.AwaitUpdateJobStatus(ctx, activities.UpdateJobStatusRequest{
			JobID:             s.JobID,
			Status:            app.RunnerJobStatusNotAttempted,
			StatusDescription: fmt.Sprintf("runner not found: %s", signal.HumanError(err)),
		})
		return pkgerrors.Wrap(err, "runner not found")
	}

	// Check if runner has any active process
	resp, err := activities.AwaitHasActiveRunnerProcess(ctx, activities.HasActiveRunnerProcessRequest{
		RunnerID: s.RunnerID,
	})
	if err != nil || !resp.HasActive {
		_ = activities.AwaitUpdateJobStatus(ctx, activities.UpdateJobStatusRequest{
			JobID:             s.JobID,
			Status:            app.RunnerJobStatusNotAttempted,
			StatusDescription: "no active runner process available",
		})
		return errors.New("runner has no active process")
	}

	// Validate job exists
	_, err = activities.AwaitGetJob(ctx, activities.GetJobRequest{ID: s.JobID})
	return pkgerrors.Wrap(err, "job not found")
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("fetching runner to ensure it is healthy")
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: s.RunnerID,
	})
	if err != nil {
		s.updateJobStatus(ctx, s.JobID, app.RunnerJobStatusNotAttempted, "unable to fetch runner from database")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	// Check if runner has any active process
	resp, err := activities.AwaitHasActiveRunnerProcess(ctx, activities.HasActiveRunnerProcessRequest{
		RunnerID: s.RunnerID,
	})
	if err != nil || !resp.HasActive {
		l.Warn("runner has no active process, not attempting")
		s.updateJobStatus(ctx, s.JobID, app.RunnerJobStatusNotAttempted, "no active runner process available")
		return errors.New("runner has no active process")
	}

	runnerJob, err := activities.AwaitGetJob(ctx, activities.GetJobRequest{
		ID: s.JobID,
	})
	if err != nil {
		s.updateJobStatus(ctx, s.JobID, app.RunnerJobStatusNotAttempted, "unable to get job from database")
		return fmt.Errorf("unable to update runner job: %w", err)
	}

	// clear any old jobs behind this that were orphaned, or not attempted for whatever reason
	if err := activities.AwaitFlushOrphanedJobs(ctx, activities.FlushOrphanedJobsRequest{
		RunnerID:  s.RunnerID,
		Threshold: runnerJob.CreatedAt.Add(-time.Minute * 5),
	}); err != nil {
		return pkgerrors.Wrap(err, "unable to flush orphaned jobs")
	}

	if runnerJob.Status == app.RunnerJobStatusCancelled {
		l.Info("job was already cancelled, not attempting")
		return nil
	}

	activities.AwaitUpdateJobStartedAt(ctx, activities.UpdateJobStartedAtRequest{JobID: s.JobID, StartedAt: workflow.Now(ctx)})
	defer func() {
		activities.AwaitUpdateJobFinishedAt(ctx, activities.UpdateJobFinishedAtRequest{JobID: s.JobID, FinishedAt: workflow.Now(ctx)})
	}()

	if runner.RunnerGroup.Settings.SandboxMode {
		l.Info("runner is in sandbox mode")
		if runner.Org.CreatedBy.AccountType == app.AccountTypeCanary {
			s.updateJobStatus(ctx, s.JobID, app.RunnerJobStatusFinished, "success in sandbox/canary mode")
			return nil
		}

		if s.cfg.SandboxModeEnableRunners {
			l.Info("enable runners enabled, so this job must be handled by a local runner",
				zap.String("job_id", s.JobID),
				zap.String("runner_id", s.RunnerID),
			)
		} else {
			l.Info("skipping job, and setting to true in sandbox mode",
				zap.String("job_id", s.JobID),
				zap.String("runner_id", s.RunnerID),
			)
			workflow.Sleep(ctx, s.cfg.SandboxModeSleep)
			s.updateJobStatus(ctx, s.JobID, app.RunnerJobStatusFinished, "success in sandbox mode")
			return nil
		}
	}

	now := workflow.Now(ctx)
	if runnerJob.CreatedAt.Add(runnerJob.QueueTimeout).Before(now) {
		s.updateJobStatus(ctx, s.JobID, app.RunnerJobStatusNotAttempted, "queue timeout reached")
		return nil
	}

	for i := 0; i < runnerJob.MaxExecutions; i++ {
		l.Info(fmt.Sprintf("attempting job execution %d of %d", i+1, runnerJob.MaxExecutions))
		retry, started, err := s.startJobExecution(ctx, runnerJob)
		if err != nil {
			return err
		}
		if !started {
			if !retry {
				return nil
			}
			continue
		}

		retry, err = s.monitorJobExecution(ctx, runnerJob)
		if err != nil {
			return err
		}
		if !retry {
			return nil
		}
	}

	return nil
}

func (s *Signal) startJobExecution(ctx workflow.Context, job *app.RunnerJob) (bool, bool, error) {
	startTS := workflow.Now(ctx)
	tags := map[string]string{
		"status":        "ok",
		"job_type":      string(job.Type),
		"job_group":     string(job.Group),
		"job_operation": string(job.Operation),
	}

	etags := maps.Clone(tags)
	etags["job_id"] = job.ID
	etags["job_operation"] = string(job.Operation)
	etags["runner_id"] = job.RunnerID
	etags["org_id"] = string(job.OrgID)
	etags["org_name"] = job.Org.Name
	etags["available_timeout"] = job.AvailableTimeout.String()
	etags["overall_timeout"] = job.OverallTimeout.String()

	defer func() {
		s.mw.Incr(ctx, "runner.job_execution_start", metrics.ToTags(tags)...)
		e2eLatency := workflow.Now(ctx).Sub(startTS)
		s.mw.Timing(ctx, "runner.job_execution_start_latency", e2eLatency, metrics.ToTags(tags)...)
	}()

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, false, err
	}

	availableStart := workflow.Now(ctx)
	l.Info("marking job as available for runner to pick up")
	s.updateJobStatus(ctx, job.ID, app.RunnerJobStatusAvailable, "waiting for runner to reserve job")

	start := workflow.Now(ctx)
	overallTimeout := job.CreatedAt.Add(job.OverallTimeout)
	availableTimeout := start.Add(job.AvailableTimeout)

	var jobExecutionFound bool
	var runnerStatus app.RunnerStatus

	if job.Group != app.RunnerJobGroupOperations {
		for attempt := 0; runnerStatus != app.RunnerStatusActive; attempt++ {
			workflow.Sleep(ctx, pollPeriod(attempt))
			etags["runner_status"] = string(runnerStatus)

			now := workflow.Now(ctx)
			if now.After(overallTimeout) {
				l.Error("overall job timeout reached")
				s.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "overall timeout waiting for runner to be healthy")
				tags["status"] = "runner_unhealthy"

				s.mw.Event(ctx, &statsd.Event{
					Title:          "Overall job timeout reached waiting for runner to become healthy",
					Text:           "Overall end-to-end job execution timeout reached while waiting for job to bewcome healthy",
					Tags:           metrics.ToTags(etags),
					SourceTypeName: "nuon-jobsys",
					Priority:       statsd.Normal,
					AlertType:      statsd.Error,
					AggregationKey: "runner-job-timeout-waiting-for-healthy-runner",
				})
				return false, false, nil
			}

			if now.After(availableTimeout) {
				l.Error("timeout waiting for job to be picked up")
				s.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "timeout waiting for runner to become healthy")
				tags["status"] = "runner_unhealthy"

				s.mw.Event(ctx, &statsd.Event{
					Title:          "Available timeout reached waiting for runner to become healthy",
					Text:           "Job is ready for execution, but runner did not become healthy within the available timeout",
					Tags:           metrics.ToTags(etags),
					SourceTypeName: "nuon-jobsys",
					Priority:       statsd.Low,
					AlertType:      statsd.Warning,
					AggregationKey: "runner-job-timeout-waiting-for-healthy-runner",
				})
				return true, false, nil
			}

			runnerStatus, err = activities.AwaitGetRunnerStatusByID(ctx, job.RunnerID)
			if err != nil {
				l.Warn("unable to determine runner status", zap.Error(err))
				return false, false, err
			}

			jobStatus, err := activities.AwaitGetJobStatusByID(ctx, job.ID)
			if err != nil {
				return false, false, nil
			}
			if jobStatus == app.RunnerJobStatusCancelled {
				l.Error("job was cancelled")
				tags["status"] = "job_cancelled"
				return false, false, nil
			}
		}
	}

	// poll until the job is picked up, and an execution exists
	for attempt := 0; !jobExecutionFound; attempt++ {
		workflow.Sleep(ctx, pollPeriod(attempt))

		now := workflow.Now(ctx)
		if now.After(overallTimeout) {
			l.Error("overall job timeout reached")
			s.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "overall timeout")
			tags["status"] = "overall_timeout"

			etags["status"] = "overall_timeout"
			s.mw.Event(ctx, &statsd.Event{
				Title:          "Overall job timeout reached without job starting",
				Text:           "Overall end-to-end job execution timeout reached without ever having been picked up",
				Tags:           metrics.ToTags(etags),
				SourceTypeName: "nuon-jobsys",
				Priority:       statsd.Normal,
				AlertType:      statsd.Error,
				AggregationKey: "runner-job-timeout-awaiting-job-pickup",
			})
			return false, false, nil
		}

		if now.After(availableTimeout) {
			l.Error("timeout waiting for job to be picked up")
			s.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "timeout waiting for runner to pick up job")
			tags["status"] = "available_timeout"

			etags["status"] = "available_timeout"
			s.mw.Event(ctx, &statsd.Event{
				Title:          "Timeout waiting for runner job to be picked up",
				Text:           "Job was marked as ready for execution, and runner appears to be in a healthy state, but runner did not pick up the job within the available timeout",
				Tags:           metrics.ToTags(etags),
				SourceTypeName: "nuon-jobsys",
				Priority:       statsd.Normal,
				AlertType:      statsd.Error,
				AggregationKey: "runner-job-timeout-awaiting-job-pickup",
			})
			return true, false, nil
		}

		jobStatus, err := activities.AwaitGetJobStatusByID(ctx, job.ID)
		if err != nil {
			return false, false, nil
		}
		if jobStatus == app.RunnerJobStatusCancelled {
			l.Error("job was cancelled")
			tags["status"] = "job_cancelled"
			return false, false, nil
		}

		jobExecutionResp, err := activities.AwaitGetLatestJobExecution(ctx, activities.GetLatestJobExecutionRequest{
			JobID:       job.ID,
			AvailableAt: availableStart,
		})
		if err != nil {
			return false, false, fmt.Errorf("error fetching latest job execution: %w", err)
		}
		jobExecutionFound = jobExecutionResp.Found
	}

	l.Info("job picked up by runner and is in progress")
	return true, true, nil
}

// executionFailureDescription returns the runner-reported error from the
// execution's StatusV2.StatusHumanDescription when present, falling back to a
// generic label (e.g. "failed") when the runner didn't report one.
func executionFailureDescription(ctx workflow.Context, jobExecutionID, fallback string) string {
	execution, err := activities.AwaitGetJobExecution(ctx, activities.GetJobExecutionRequest{
		JobExecutionID: jobExecutionID,
	})
	if err != nil || execution == nil {
		return fallback
	}
	if desc := execution.StatusV2.StatusHumanDescription; desc != "" {
		return desc
	}
	return fallback
}

func (s *Signal) monitorJobExecution(ctx workflow.Context, job *app.RunnerJob) (bool, error) {
	startTS := workflow.Now(ctx)
	tags := map[string]string{
		"status":    "ok",
		"job_type":  string(job.Type),
		"job_group": string(job.Group),
	}
	etags := maps.Clone(tags)
	etags["job_id"] = job.ID
	etags["job_operation"] = string(job.Operation)
	etags["runner_id"] = job.RunnerID
	etags["org_id"] = string(job.OrgID)
	etags["org_name"] = job.Org.Name
	etags["available_timeout"] = job.AvailableTimeout.String()
	etags["overall_timeout"] = job.OverallTimeout.String()

	defer func() {
		s.mw.Incr(ctx, "runner.job_execution", metrics.ToTags(tags)...)
		e2eLatency := workflow.Now(ctx).Sub(startTS)
		s.mw.Timing(ctx, "runner.job_execution.latency", e2eLatency, metrics.ToTags(tags)...)
	}()

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
	}

	jobExecution, err := activities.AwaitGetCurrentJobExecutionByJobID(ctx, job.ID)
	if err != nil {
		return false, fmt.Errorf("error fetching latest job execution: %w", err)
	}

	// poll the job execution, until it's completed
	executionTimeout := jobExecution.CreatedAt.Add(job.ExecutionTimeout)
	for attempt := 0; ; attempt++ {
		workflow.Sleep(ctx, pollPeriod(attempt))

		now := workflow.Now(ctx)

		// when the overall timeout is hit, we mark both the runner job and execution as timed out
		// this is not retryable.
		if now.After(job.CreatedAt.Add(job.OverallTimeout)) {
			l.Error("overall timeout reached")
			s.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "overall timeout")
			s.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusTimedOut)
			tags["status"] = "overall_timeout"

			etags["job_status"] = string(app.RunnerJobStatusTimedOut)
			maps.Copy(etags, tags)
			s.mw.Event(ctx, &statsd.Event{
				Title:          "Overall timeout reached while job executing",
				Text:           "Overall end-to-end job execution timeout reached while waiting for job to bewcome healthy",
				Tags:           metrics.ToTags(etags),
				SourceTypeName: "nuon-jobsys",
				Priority:       statsd.Normal,
				AlertType:      statsd.Error,
				AggregationKey: "runner-job-timeout-while-executing",
			})
			return false, nil
		}

		// when the execution timeout is hit, we mark both the runner job and execution as timed out
		// this is retryable
		if now.After(executionTimeout) {
			l.Error("execution timeout reached")
			s.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "execution timeout")
			s.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusTimedOut)
			tags["status"] = "execution_timeout"

			etags["job_status"] = string(app.RunnerJobStatusTimedOut)
			maps.Copy(etags, tags)
			s.mw.Event(ctx, &statsd.Event{
				Title:          "Overall timeout reached while job executing",
				Text:           "Overall end-to-end job execution timeout reached while waiting for job to become healthy",
				Tags:           metrics.ToTags(etags),
				SourceTypeName: "nuon-jobsys",
				Priority:       statsd.Normal,
				AlertType:      statsd.Error,
				AggregationKey: "runner-job-timeout-while-executing",
			})
			return true, nil
		}

		// if the runner was started after this execution was created, we mark the execution as in error
		// this is retryable
		hb, err := activities.AwaitGetMostRecentHeartBeatRequestByRunnerID(ctx, job.RunnerID)
		if err != nil {
			return false, err
		}
		if hb == nil {
			return false, errors.New("no heart beats found")
		}

		// if the runner is restarted, we want to add a buffer before canceling any jobs in flight
		maxAliveTime := jobExecution.CreatedAt.Add(time.Minute)
		if hb.StartedAt.After(maxAliveTime) {
			l.Error(
				"runner restarted while job was in flight. job will be cancelled.",
				zap.Time("runner.started_at", hb.StartedAt),
				zap.Time("job_execution.created_at", jobExecution.CreatedAt),
			)
			s.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusCancelled)

			maps.Copy(etags, tags)
			s.mw.Event(ctx, &statsd.Event{
				Title:          "Runner restarted while job in flight",
				Text:           "A runner was marked unhealthy during the job execution. The job will NOT be resumed if/when the runner recovers",
				Tags:           metrics.ToTags(etags),
				SourceTypeName: "nuon-jobsys",
				Priority:       statsd.Normal,
				AlertType:      statsd.Error,
				AggregationKey: "runner-job-dropped",
			})
		}

		jobStatus, err := activities.AwaitGetJobStatusByID(ctx, job.ID)
		if err != nil {
			return false, err
		}
		if jobStatus == app.RunnerJobStatusCancelled {
			l.Error("job was cancelled")
			s.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusCancelled)
			tags["status"] = "cancelled"
			return true, nil
		}

		// if the runner has no active process, the job execution is marked as failed.
		processResp, err := activities.AwaitHasActiveRunnerProcess(ctx, activities.HasActiveRunnerProcessRequest{
			RunnerID: job.RunnerID,
		})
		if err != nil || !processResp.HasActive {
			l.Error("runner marked unhealthy during job")
			s.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, "runner became unhealthy during job")
			s.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusFailed)
			tags["status"] = "runner_unhealthy"

			maps.Copy(etags, tags)
			s.mw.Event(ctx, &statsd.Event{
				Title:          "Runner marked unhealthy during job",
				Text:           "A runner was marked unhealthy during the job execution. The job will NOT be resumed if/when the runner recovers",
				Tags:           metrics.ToTags(etags),
				SourceTypeName: "nuon-jobsys",
				Priority:       statsd.Normal,
				AlertType:      statsd.Error,
				AggregationKey: "runner-job-dropped",
			})
			return true, nil
		}

		executionStatus, err := activities.AwaitGetJobExecutionStatus(ctx, activities.GetJobExecutionStatusRequest{
			JobExecutionID: jobExecution.ID,
		})
		if err != nil {
			return false, err
		}

		switch executionStatus {
		case app.RunnerJobExecutionStatusFinished:
			l.Info("job execution successfully finished")
			s.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFinished, "finished")
			tags["status"] = "ok"
			return false, nil
		case app.RunnerJobExecutionStatusCancelled:
			l.Info("job cancelled")
			tags["status"] = "execution_cancelled"
			return true, nil
		case app.RunnerJobExecutionStatusFailed:
			l.Info("job execution failed")
			s.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, executionFailureDescription(ctx, jobExecution.ID, "failed"))
			tags["status"] = "execution_failed"
			return true, nil
		case app.RunnerJobExecutionStatusTimedOut:
			l.Info("job execution timed out")
			s.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, executionFailureDescription(ctx, jobExecution.ID, "execution timed out"))
			tags["status"] = "execution_timed_out"
			return true, nil
		case app.RunnerJobExecutionStatusNotAttempted:
			l.Info("job execution not attempted")
			s.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, executionFailureDescription(ctx, jobExecution.ID, "execution not attempted"))
			tags["status"] = "execution_not_attempted"
			return true, nil
		default:
			continue
		}
	}
}

func (s *Signal) updateJobStatus(ctx workflow.Context, jobID string, status app.RunnerJobStatus, statusDescription string) {
	err := activities.AwaitUpdateJobStatus(ctx, activities.UpdateJobStatusRequest{
		JobID:             jobID,
		Status:            status,
		StatusDescription: statusDescription,
	})

	statusactivities.AwaitUpdateRunnerJobStatusV2(ctx, statusactivities.UpdateRunnerJobStatusV2Request{
		RunnerJobID:       jobID,
		Status:            status,
		StatusDescription: statusDescription,
	})

	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update runner job status",
		zap.String("runner-job-id", jobID),
		zap.Error(err))
}

func (s *Signal) updateJobExecutionStatus(ctx workflow.Context, jobExecutionID string, status app.RunnerJobExecutionStatus) {
	err := activities.AwaitUpdateJobExecutionStatus(ctx, activities.UpdateJobExecutionStatusRequest{
		JobExecutionID: jobExecutionID,
		Status:         status,
	})

	statusactivities.AwaitUpdateRunnerJobExecutionStatusV2(ctx, statusactivities.UpdateRunnerJobExecutionStatusV2Request{
		RunnerJobExecutionID: jobExecutionID,
		Status:               status,
	})

	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update runner job execution status",
		zap.String("runner-job-execution id", jobExecutionID),
		zap.Error(err))
}
