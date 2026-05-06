package appconfigupdated

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/driftchecksandbox"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

// reconcileDriftSandboxEmitter reconciles a single per-install sandbox drift
// cron emitter against AppSandboxConfig.DriftSchedule. Mirrors
// reconcileDriftEmitters but install-scoped (sandbox drift is configured at
// the app-sandbox-config level, not per-component).
//
// `existing` is the pre-filtered list of emitters whose name matched the
// driftSandboxEmitterPrefix. They're stopped and deleted unconditionally;
// when the schedule is non-empty a fresh emitter replaces them.
func (s *Signal) reconcileDriftSandboxEmitter(
	ctx workflow.Context,
	l log.Logger,
	install *app.Install,
	queue *app.Queue,
	existing []app.QueueEmitter,
) error {
	stopAndDeleteEmitters(ctx, l, existing)

	schedule := install.AppSandboxConfig.DriftSchedule
	if schedule == "" {
		return nil
	}

	name := driftSandboxEmitterPrefix + install.ID
	if _, err := emitterclient.AwaitCreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:      queue.ID,
		Name:         name,
		Description:  fmt.Sprintf("sandbox drift check for install %s", install.ID),
		Mode:         app.QueueEmitterModeCron,
		CronSchedule: schedule,
		SignalType:   driftchecksandbox.SignalType,
		SignalTemplate: &driftchecksandbox.Signal{
			InstallID: install.ID,
		},
	}); err != nil {
		return fmt.Errorf("unable to create sandbox drift emitter %s: %w", name, err)
	}

	return nil
}
