package worker

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// Check if there was a recent restart, and force one if the runner has been running for
// longer than the configured time.

type HealthcheckUpdateNeededRequest struct {
	HeartbeatID string `validate:"required"`
	RunnerID    string `validate:"required"`
}

type HealthcheckUpdateNeededResponse struct {
	ShouldUpdate bool
}

func HealthcheckUpdateNeededWorkflowsID(req *HealthcheckUpdateNeededRequest) string {
	return fmt.Sprintf("healthcheck-update-needed-%s", req.RunnerID)
}

// @temporal-gen workflow
// @execution-timeout 2m
// @task-timeout 5m
// @id-callback HealthcheckUpdateNeededWorkflowsID
func (w *Workflows) HealthcheckUpdateNeeded(
	ctx workflow.Context,
	req *HealthcheckUpdateNeededRequest,

) (*HealthcheckUpdateNeededResponse, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return &HealthcheckUpdateNeededResponse{}, errors.Wrap(err, "unable to get workflow logger")
	}
	runner, err := activities.AwaitGetByRunnerID(ctx, req.RunnerID)
	if err != nil {
		return &HealthcheckUpdateNeededResponse{}, errors.Wrap(err, "unable to get runner by id")
	}

	heartbeat, err := activities.AwaitGetHeartBeatByID(ctx, req.HeartbeatID)
	if err != nil {

		return &HealthcheckUpdateNeededResponse{}, errors.Wrap(err, "unable to get heartbeat by id")
	}
	var needsUpdate bool
	if runner.RunnerGroup.Settings.ExpectedVersion == "latest" {
		needsUpdate = heartbeat.Version != w.cfg.Version
	} else if heartbeat.Version != runner.RunnerGroup.Settings.ExpectedVersion {
		// NOTE(sdboyer) this branch is unreachable until we have a versioning
		// strategy other than latest.
		//
		// However, a lot of older orgs _do_ have something other than `latest`
		// set for their expected version, and as a result they're looping here.
		// So we just never update these, until we have a better strategy.
		needsUpdate = false

	}

	// NOTE(jm): if we need an update, we just write a metric
	w.mw.Incr(ctx, "runner.version_update", metrics.ToTags(map[string]string{
		"runner_type":          string(runner.RunnerGroup.Type),
		"needs_version_update": strconv.FormatBool(needsUpdate),
		"expected_latest":      strconv.FormatBool(runner.RunnerGroup.Settings.ExpectedVersion == "latest"),
	})...)

	if needsUpdate {
		l.Info("sending signal to update out-of-date runner",
			zap.String("runner_id", runner.ID),
			zap.String("runner_type", string(runner.RunnerGroup.Type)),
			zap.String("expected_version", runner.RunnerGroup.Settings.ExpectedVersion),
			zap.String("reported_version", heartbeat.Version),
			zap.String("api_version", w.cfg.Version),
		)

		//w.evClient.Send(ctx, runnerID, &signals.RequestSignal{
		//Signal: &signals.Signal{
		//Type:          signals.OperationUpdateVersion,
		//HealthCheckID: healthcheck.ID,
		//},
		//EventLoopRequest: eventloop.EventLoopRequest{
		//ID: runnerID,
		//},
		//})
	}

	return &HealthcheckUpdateNeededResponse{ShouldUpdate: needsUpdate}, nil
}
