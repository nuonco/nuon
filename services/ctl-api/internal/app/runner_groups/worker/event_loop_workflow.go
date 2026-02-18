package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runner_groups/signals"
	runnersactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/loop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationElectLeader: w.handleElectLeader,
		signals.OperationSetLeader:   w.handleSetLeader,
	}

	l := loop.Loop[*signals.Signal, signals.RequestSignal]{
		Cfg:              w.cfg,
		V:                w.v,
		MW:               w.mw,
		Handlers:         handlers,
		NewRequestSignal: signals.NewRequestSignal,
		ExistsHook: func(ctx workflow.Context, req eventloop.EventLoopRequest) (bool, error) {
			return true, nil
		},
	}

	return l.Run(ctx, req, pendingSignals)
}

func (w *Workflows) handleSetLeader(ctx workflow.Context, req signals.RequestSignal) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	result, err := runnersactivities.AwaitSetLeader(ctx, runnersactivities.SetLeaderRequest{
		RunnerGroupID: req.ID,
		RunnerID:      req.RequestedLeaderRunnerID,
	})
	if err != nil {
		l.Error("set leader failed",
			zap.String("runner_group_id", req.ID),
			zap.String("requested_runner_id", req.RequestedLeaderRunnerID),
			zap.Error(err),
		)
		return err
	}

	// If leadership changed, reschedule queued jobs from old to new leader.
	if result.NewLeaderID != "" && result.OldLeaderID != "" && result.OldLeaderID != result.NewLeaderID {
		if _, err := runnersactivities.AwaitRescheduleJobsToLeader(ctx, runnersactivities.RescheduleJobsToLeaderRequest{
			OldLeaderRunnerID: result.OldLeaderID,
			NewLeaderRunnerID: result.NewLeaderID,
		}); err != nil {
			l.Error("unable to reschedule jobs to new leader",
				zap.String("old_leader", result.OldLeaderID),
				zap.String("new_leader", result.NewLeaderID),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (w *Workflows) handleElectLeader(ctx workflow.Context, req signals.RequestSignal) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	result, err := runnersactivities.AwaitElectLeader(ctx, runnersactivities.ElectLeaderRequest{
		RunnerGroupID: req.ID,
	})
	if err != nil {
		l.Error("leader election failed",
			zap.String("runner_group_id", req.ID),
			zap.Error(err),
		)
		return err
	}

	// If leadership changed, reschedule queued jobs from old to new leader.
	if result.NewLeaderID != "" && result.OldLeaderID != "" && result.OldLeaderID != result.NewLeaderID {
		if _, err := runnersactivities.AwaitRescheduleJobsToLeader(ctx, runnersactivities.RescheduleJobsToLeaderRequest{
			OldLeaderRunnerID: result.OldLeaderID,
			NewLeaderRunnerID: result.NewLeaderID,
		}); err != nil {
			l.Error("unable to reschedule jobs to new leader",
				zap.String("old_leader", result.OldLeaderID),
				zap.String("new_leader", result.NewLeaderID),
				zap.Error(err),
			)
		}
	}

	return nil
}
