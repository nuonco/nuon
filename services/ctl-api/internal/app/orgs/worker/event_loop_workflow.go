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
	l := workflow.GetLogger(ctx)

	defaultTags := map[string]string{"sandbox_mode": strconv.FormatBool(req.SandboxMode)}
	w.mw.Incr(ctx, "event_loop.start", metrics.ToTags(defaultTags, "op", "started")...)

	finished := false
	signalChan := workflow.GetSignalChannel(ctx, req.ID)
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, _ bool) {
		var evSignal sigs.Signal
		channelOpen := channel.Receive(ctx, &evSignal)
		if !channelOpen {
			l.Info("channel was closed")
			return
		} else {
			l.Info("channel is open")
		}

		//if err := evSignal.Validate(w.v); err != nil {
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

		sreq := sigs.RequestSignal{
			Signal:           &evSignal,
			EventLoopRequest: req,
		}

		switch evSignal.SignalType() {
		case sigs.OperationCreated:
			op = "created"
			err := w.AwaitCreated(ctx, sreq)
			if err != nil {
				status = "error"
				l.Error("unable to handle created signal", zap.Error(err))
				return
			}
		case sigs.OperationProvision:
			op = "provision"
			err := w.AwaitProvision(ctx, sreq)
			if err != nil {
				status = "error"
				l.Error("unable to provision org", zap.Error(err))
				return
			}
		case sigs.OperationReprovision:
			op = "reprovision"
			err := w.AwaitReprovision(ctx, sreq)
			if err != nil {
				status = "error"
				l.Error("unable to reprovision org", zap.Error(err))
				return
			}
		case sigs.OperationDeprovision:
			op = "deprovision"
			err := w.AwaitDeprovision(ctx, sreq)
			if err != nil {
				status = "error"
				l.Error("unable to deprovision org", zap.Error(err))
				return
			}
		case sigs.OperationForceDeprovision:
			op = "force_deprovision"
			err := w.AwaitForceDeprovision(ctx, sreq)
			if err != nil {
				status = "error"
				l.Error("unable to force deprovision org", zap.Error(err))
				return
			}
		case sigs.OperationRestart:
			op = "restart"
		case sigs.OperationInviteCreated:
			op = "invite_created"
			err := w.AwaitInviteUser(ctx, sreq)
			if err != nil {
				status = "error"
				l.Error("unable to handle invite created signal", zap.Error(err))
				return
			}
		case sigs.OperationInviteAccepted:
			op = "invite_accepted"
			err := w.AwaitInviteAccepted(ctx, sreq)
			if err != nil {
				status = "error"
				l.Error("unable to handle invite accepted signal", zap.Error(err))
				return
			}
		case sigs.OperationForceDelete:
			op = "force_delete"
			err := w.AwaitForceDelete(ctx, sreq)
			if err != nil {
				status = "error"
				l.Error("unable to force delete org", zap.Error(err))
				return
			}

			finished = true
		case sigs.OperationDelete:
			op = "delete"
			err := w.AwaitDelete(ctx, sreq)
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
