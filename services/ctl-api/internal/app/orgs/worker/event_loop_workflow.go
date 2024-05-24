package worker

import (
	"errors"
	"fmt"
	"strconv"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/signals"
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
	w.mw.Incr(ctx, "event_loop.start", metrics.ToTags(defaultTags, "op", "started")...)
	l := workflow.GetLogger(ctx)

	finished := false
	signalChan := workflow.GetSignalChannel(ctx, req.OrgID)
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, _ bool) {
		var signal signals.Signal
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
			w.mw.Incr(ctx, "event_loop.signal", metrics.ToTags(tags)...)
		}()

		switch signal.Operation {
		// OperationCreated
		case signals.OperationCreated:
			op = "created"
			err := w.created(ctx, req.OrgID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to handle created signal", zap.Error(err))
				return
			}
		// OperationProvision
		case signals.OperationProvision:
			op = "provision"
			err := w.provision(ctx, req.OrgID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to provision org", zap.Error(err))
				return
			}

		// OperationReprovision
		case signals.OperationReprovision:
			op = "reprovision"
			err := w.reprovision(ctx, req.OrgID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to reprovision org", zap.Error(err))
				return
			}

		// OperationDeprovision
		case signals.OperationDeprovision:
			op = "deprovision"
			err := w.deprovision(ctx, req.OrgID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to deprovision org", zap.Error(err))
				return
			}

		// OperationRestart
		case signals.OperationRestart:
			op = "restart"
			w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
				OrgID:       req.OrgID,
				SandboxMode: req.SandboxMode,
			})

			// OperationInviteCreated
		case signals.OperationInviteCreated:
			op = "invite_created"
			err := w.inviteUser(ctx, req.OrgID, signal.InviteID)
			if err != nil {
				status = "error"
				l.Error("unable to handle invite created signal", zap.Error(err))
				return
			}

			// OperationInviteAccepted
		case signals.OperationInviteAccepted:
			op = "invite_accepted"
			err := w.inviteAccepted(ctx, req.OrgID, signal.InviteID)
			if err != nil {
				status = "error"
				l.Error("unable to handle invite accepted signal", zap.Error(err))
				return
			}

			// OperationForceDelete
		case signals.OperationForceDelete:
			op = "force_delete"
			err := w.forceDelete(ctx, req.OrgID, req.SandboxMode)
			if err != nil {
				status = "error"
				l.Error("unable to force delete org", zap.Error(err))
				return
			}

			finished = true

		// OperationDelete
		case signals.OperationDelete:
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
			w.mw.Incr(ctx, "event_loop.canceled", metrics.ToTags(defaultTags)...)
			l.Error("workflow canceled")
			break
		}

		selector.Select(ctx)
	}

	return nil
}
