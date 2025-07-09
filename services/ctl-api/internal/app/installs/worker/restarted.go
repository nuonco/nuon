package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 1m
func (w *Workflows) Restarted(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}
	return w.ensureSubLoops(ctx, install, sreq)
}

func (w *Workflows) ensureSubLoops(ctx workflow.Context, install *app.Install, sreq signals.RequestSignal) error {
	// stack, err := activities.AwaitGetInstallStackByInstallID(ctx, sreq.ID)
	// if err != nil && !strings.Contains(err.Error(), "record not found") {
	// 	return errors.Wrap(err, "unable to get install stack")
	// }

	// older installs may not have a stack
	if install.InstallStack.ID != "" {
		sreq.ID = fmt.Sprintf("%s-%s-%s", install.ID, "stack", install.InstallStack.ID)
		sreq.EventLoopWorkflowType = "StackEventLoop"

		_, err := w.evClient.SendAsync(ctx, sreq.ID, &sreq)
		if err != nil {
			return errors.Wrapf(err, "unable to send restart signal to stack event loop %s", sreq.ID)
		}
		sreq.SignalListeners = nil
	}

	{
		sreq.ID = fmt.Sprintf("%s-%s-%s", install.ID, "sandbox", install.InstallSandbox.ID)
		sreq.EventLoopWorkflowType = "SandboxEventLoop"

		_, err := w.evClient.SendAsync(ctx, sreq.ID, &sreq)
		if err != nil {
			return errors.Wrapf(err, "unable to send restart signal to sandbox event loop %s", sreq.ID)
		}
		sreq.SignalListeners = nil
	}

	componentIDs, err := activities.AwaitGetInstallComponentIDsByInstallID(ctx, install.ID)
	if err != nil {
		return err
	}
	for _, id := range componentIDs {
		sreq.ID = fmt.Sprintf("%s-%s-%s", install.ID, "component", id)
		sreq.EventLoopWorkflowType = "ComponentEventLoop"

		_, err := w.evClient.SendAsync(ctx, sreq.ID, &sreq)
		if err != nil {
			return errors.Wrapf(err, "unable to send restart signal to install component event loop %s", sreq.ID)
		}
		sreq.SignalListeners = nil
	}

	iaws, err := activities.AwaitGetActionWorkflowsByInstallID(ctx, install.ID)
	if err != nil {
		return err
	}
	for _, iaw := range iaws {
		sreq.ID = fmt.Sprintf("%s-%s-%s", install.ID, "action", iaw.ID)
		sreq.EventLoopWorkflowType = "ActionEventLoop"

		_, err := w.evClient.SendAsync(ctx, sreq.ID, &sreq)
		if err != nil {
			return errors.Wrapf(err, "unable to send restart signal to install action workflow event loop %s", sreq.ID)
		}
		sreq.SignalListeners = nil
	}

	return nil
}
