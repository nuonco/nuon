package worker

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	EventLoopWorkflowName string = "OrgEventLoop"
	defaultOrgRegion      string = "us-west-2"
)

func EventLoopWorkflowID(orgID string) string {
	return fmt.Sprintf("%s-event-loop", orgID)
}

type OrgEventLoopRequest struct {
	OrgID       string
	SandboxMode bool
}

func (w *Workflows) OrgEventLoop(ctx workflow.Context, req OrgEventLoopRequest) error {
	defaultTags := map[string]string{"sandbox_mode": strconv.FormatBool(req.SandboxMode)}
	w.mw.Incr(ctx, "event_loop.start", 1, metrics.ToTags(defaultTags, "op", "started")...)
	l := workflow.GetLogger(ctx)

	finished := false
	signalChan := workflow.GetSignalChannel(ctx, req.OrgID)
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, _ bool) {
		var signal Signal
		channelOpen := channel.Receive(ctx, &signal)
		if !channelOpen {
			l.Info("channel was closed")
			return
		}

		if err := signal.Validate(w.v); err != nil {
			l.Info("invalid signal", zap.Error(err))
		}

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
			w.mw.Incr(ctx, "event_loop.signal", 1, metrics.ToTags(tags)...)
		}()

		switch signal.Operation {
		// OperationProvision
		case OperationProvision:
			op = "provision"
			err := w.provision(ctx, req.OrgID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to provision org", zap.Error(err))
				return
			}

		// OperationReprovision
		case OperationReprovision:
			op = "reprovision"
			err := w.reprovision(ctx, req.OrgID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to reprovision org", zap.Error(err))
				return
			}

		// OperationDeprovision
		case OperationDeprovision:
			op = "deprovision"
			err := w.deprovision(ctx, req.OrgID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to deprovision org", zap.Error(err))
				return
			}

		// OperationRestart
		case OperationRestart:
			op = "restart"
			w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
				OrgID:       req.OrgID,
				SandboxMode: req.SandboxMode,
			})

		// OperationDelete
		case OperationDelete:
			op = "delete"
			err := w.delete(ctx, req.OrgID, req.SandboxMode)
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
			w.mw.Incr(ctx, "event_loop.canceled", 1, metrics.ToTags(defaultTags)...)
			l.Error("workflow canceled")
			break
		}

		if temporal.IsPanicError(ctx.Err()) {
			w.mw.Incr(ctx, "event_loop.panic", 1, metrics.ToTags(defaultTags)...)
			w.mw.Event(ctx, &statsd.Event{
				Title: "event loop panic",
				Text:  "event loop panic\n\t-" + req.OrgID,
			})
			l.Error("workflow panic", zap.Error(ctx.Err()))
			break
		}

		selector.Select(ctx)
	}

	return nil
}
