package worker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type HealthcheckJobRunnerRequest struct {
	HealthCheckID string `validate:"required"`
	RunnerID      string `validate:"required"`
}

type HealthcheckJobRunnerResponse struct {
	ShouldRestart bool `json:"should_restart,omitzero"`
}

func HealthcheckJobRunnerWorkflowsID(req *HealthcheckJobRunnerRequest) string {
	return fmt.Sprintf("healthcheck-job-%s", req.RunnerID)
}

// @temporal-gen workflow
// @execution-timeout 3m
// @task-timeout 5m
// @id-callback HealthcheckJobRunnerWorkflowsID
func (w *Workflows) HealthcheckJobRunner(ctx workflow.Context, req *HealthcheckJobRunnerRequest) (*HealthcheckJobRunnerResponse, error) {
	l, err := log.WorkflowLogger(ctx)
	logStream, err := activities.AwaitCreateLogStreamByOperationID(ctx, req.HealthCheckID)
	if err != nil {
		return &HealthcheckJobRunnerResponse{ShouldRestart: false}, errors.Wrap(err, "unable to create logstream")
	}
	// create a runner job
	job, err := activities.AwaitCreateHealthCheckJob(
		ctx,
		&activities.CreateHealthCheckJobRequest{
			RunnerID:    req.RunnerID,
			OwnerID:     req.HealthCheckID, // should this be healthcheck id
			LogStreamID: logStream.ID,
			Metadata:    map[string]string{},
		},
	)
	if err != nil {
		return &HealthcheckJobRunnerResponse{ShouldRestart: false}, errors.Wrap(err, "unable to create health check job")
	}

	// create the signal request
	request := &signals.RequestSignal{
		Signal: &signals.Signal{
			Type:          signals.OperationHealthcheck,
			HealthCheckID: req.HealthCheckID,
			JobID:         job.ID,
		},
		EventLoopRequest: eventloop.EventLoopRequest{
			ID: req.RunnerID,
		},
	}

	// NOTE(fd): this should always fire even if the are other jobs in flight/executing.
	w.AwaitProcessJob(ctx, *request)
	workflow.Sleep(ctx, time.Duration(5)*time.Second) // sleep so the outputs "flush"

	// get the job's execution and the outputs
	// TODO(fd): write a helper to get the outputs from the latest job execution
	job, err = activities.AwaitGetJobByID(ctx, job.ID)
	if err != nil {
		return &HealthcheckJobRunnerResponse{ShouldRestart: false}, errors.Wrap(err, "unable to get job execution")
	}

	// determine if we should restart
	shouldRestart := false

	outputs := job.ParsedOutputs
	fmt.Println(outputs)
	hc := &configs.HealthcheckOutputs{}
	b, err := json.Marshal(outputs)
	json.Unmarshal(b, hc)

	// check the job loops Healthcheck ouptuts if we have missed one of these runner-side healthchecks,
	// the values in the payload will be larger than  the interval at which this job is run. if that's
	// the case, we should restart.
	zeroCount := 0
	for job_loop, dur := range hc.JobLoops {
		if dur == 0 {
			zeroCount += 1
			l.Debug("job loop time since last healthcheck is zero",
				zap.String("job_loop", job_loop),
				zap.String("duration", dur.String()),
			)
		} else if dur > (runnerSideCheckInterval * 2) { // likely missed one healthcheck: is this enough to merit a restart?
			l.Info("runner should restart on the basis of job loop last healthcheck being more than double the expected interval",
				zap.String("job_loop", job_loop),
				zap.String("duration", dur.String()),
			)
			shouldRestart = true
		}
	}
	// NOTE(fd): this case seems very unlikely
	if zeroCount > 0 && zeroCount != len(hc.JobLoops) {
		// some job loops may have failed to start
		l.Debug("Some job loops may have failed to start.")
		if !shouldRestart {
			shouldRestart = true
		}
	}

	return &HealthcheckJobRunnerResponse{ShouldRestart: shouldRestart}, nil
}
