package worker

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// Check if there was a recent restart, and force one if the runner has been running for
// longer than the configured time.

type HealthcheckCheckRestartRequest struct {
	HeartbeatID string `validate:"required"`
	RunnerID    string `validate:"required"`
}

type HealthcheckCheckRestartResponse struct {
	ShouldRestart bool
}

func HealthcheckCheckRestartWorkflowsID(req *HealthcheckCheckRestartRequest) string {
	return fmt.Sprintf("healthcheck-check-restart-%s", req.RunnerID)
}

// @temporal-gen workflow
// @execution-timeout 3m
// @task-timeout 5m
// @id-callback HealthcheckCheckRestartWorkflowsID
func (w *Workflows) HealthcheckCheckRestart(
	ctx workflow.Context,
	req *HealthcheckCheckRestartRequest,
) (*HealthcheckCheckRestartResponse, error) {
	runner, err := activities.AwaitGetByRunnerID(ctx, req.RunnerID)
	if err != nil {
		return &HealthcheckCheckRestartResponse{}, errors.Wrap(err, "unable to get runner by id")
	}

	heartbeat, err := activities.AwaitGetHeartBeatByID(ctx, req.HeartbeatID)
	if err != nil {
		return &HealthcheckCheckRestartResponse{}, errors.Wrap(err, "unable to get heartbeat by id")
	}

	w.mw.Gauge(ctx, "runner.alivetime", float64(heartbeat.AliveTime.Seconds()), metrics.ToTags(map[string]string{
		"runner_type": string(runner.RunnerGroup.Type),
	})...)

	// TODO(sdboyer) replace with actual value from group settings when actually implementing the call
	// TODO(sdboyer) this is artificially low for testing purposes
	ttl := time.Minute * 10
	if heartbeat.AliveTime < time.Second*5 {
		w.mw.Incr(ctx, "runner.restart", metrics.ToTags(map[string]string{
			"runner_type": string(runner.RunnerGroup.Type),
		})...)
	} else if heartbeat.AliveTime > ttl {
		w.mw.Incr(ctx, "runner.ttl_exceeded", metrics.ToTags(map[string]string{
			"runner_type": string(runner.RunnerGroup.Type),
			"ttl":         ttl.String(),
			"alive_for":   heartbeat.AliveTime.String(),
		})...)

		// TODO(sdboyer) temporary for more granular clarity than the metric gives. Remove once restart is implemented
		l, err := log.WorkflowLogger(ctx)
		if err != nil {
			return &HealthcheckCheckRestartResponse{ShouldRestart: false}, err
		}
		l.Info("runner ttl exceeded, scheduling restart",
			zap.String("runner_id", runner.ID),
			zap.String("runner_type", string(runner.RunnerGroup.Type)),
			zap.String("ttl", runner.RunnerGroup.Settings.ExpectedVersion),
			zap.String("alive_for", heartbeat.Version),
		)

		// TODO(sdboyer) actually dispatch a call in addition to the telemetry
		return &HealthcheckCheckRestartResponse{ShouldRestart: true}, nil
	}

	return &HealthcheckCheckRestartResponse{ShouldRestart: false}, nil
}
