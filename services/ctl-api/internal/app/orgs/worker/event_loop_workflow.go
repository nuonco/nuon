package worker

import (
	"errors"
	"strconv"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

const (
	defaultOrgRegion string = "us-west-2"
)

type OrgEventLoopRequest struct {
	OrgID       string
	SandboxMode bool
}

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest) error {
	defaultTags := map[string]string{"sandbox_mode": strconv.FormatBool(req.SandboxMode)}
	w.mw.Incr(ctx, "event_loop.start", metrics.ToTags(defaultTags, "op", "started")...)
	l := workflow.GetLogger(ctx)

	finished := false
	signalChan := workflow.GetSignalChannel(ctx, req.ID)
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, _ bool) {
		var evSignal sigs.Signal
		channelOpen := channel.Receive(ctx, &evSignal)
		if !channelOpen {
			l.Info("channel was closed")
			return
		}

		//if err := signal.Validate(w.v); err != nil {
		//l.Info("invalid signal", zap.Error(err))
		//}

		startTS := workflow.Now(ctx)
		op := ""
		status := "ok"
		defer func() {
			tags := generics.MergeMap(map[string]string{
				"op":     op,
				"status": status,
			}, defaultTags)
			dur := workflow.Now(ctx).Sub(startTS)

			w.mw.Timing(ctx, "event_loop.signal_duration", dur, metrics.ToTags(tags)...)
			w.mw.Incr(ctx, "event_loop.signal", metrics.ToTags(tags)...)
		}()

		switch evSignal.SignalType() {
		// OperationCreated
		case sigs.OperationCreated:
			op = "created"
			err := w.created(ctx, req.ID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to handle created signal", zap.Error(err))
				return
			}
		// OperationProvision
		case sigs.OperationProvision:
			op = "provision"
			err := w.provision(ctx, req.ID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to provision org", zap.Error(err))
				return
			}

		// OperationReprovision
		case sigs.OperationReprovision:
			op = "reprovision"
			err := w.reprovision(ctx, req.ID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to reprovision org", zap.Error(err))
				return
			}

		// OperationDeprovision
		case sigs.OperationDeprovision:
			op = "deprovision"
			err := w.deprovision(ctx, req.ID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to deprovision org", zap.Error(err))
				return
			}

		// OperationRestart
		case sigs.OperationRestart:
			op = "restart"
			w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
				OrgID:       req.ID,
				SandboxMode: req.SandboxMode,
			})

			// OperationInviteCreated
		case sigs.OperationInviteCreated:
			op = "invite_created"
			err := w.inviteUser(ctx, req.ID, evSignal.InviteID)
			if err != nil {
				status = "error"
				l.Error("unable to handle invite created signal", zap.Error(err))
				return
			}

			// OperationInviteAccepted
		case sigs.OperationInviteAccepted:
			op = "invite_accepted"
			err := w.inviteAccepted(ctx, req.ID, evSignal.InviteID)
			if err != nil {
				status = "error"
				l.Error("unable to handle invite accepted signal", zap.Error(err))
				return
			}

			// OperationForceDelete
		case sigs.OperationForceDelete:
			op = "force_delete"
			err := w.forceDelete(ctx, req.ID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to force delete org", zap.Error(err))
				return
			}

			finished = true

		// OperationDelete
		case sigs.OperationDelete:
			op = "delete"
			err := w.delete(ctx, req.ID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to delete org", zap.Error(err))
				return
			}

			finished = true
		}
	})
	for !finished {
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			w.mw.Incr(ctx, "event_loop.canceled", metrics.ToTags(defaultTags)...)
			l.Error("workflow canceled")
			break
		}

		selector.Select(ctx)
	}

	return nil
}
