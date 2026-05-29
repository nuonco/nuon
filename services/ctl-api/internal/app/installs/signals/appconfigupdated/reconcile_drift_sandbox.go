package appconfigupdated

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/driftchecksandbox"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

// reconcileDriftSandboxEmitter reconciles a single per-install sandbox drift
// cron emitter against the sandbox drift schedule. Mirrors reconcileDriftEmitters
// but install-scoped (sandbox drift is configured at the app-sandbox-config level,
// not per-component).
//
// The schedule is read from appCfg.SandboxConfig (the install's CURRENT app
// config, which `apps sync` advances via install.AppConfigID) — NOT from
// install.AppSandboxConfig, whose pinned AppSandboxConfigID is not advanced on
// sync and would therefore reconcile against a stale (or no-longer-disabled)
// schedule.
//
// `existing` is the pre-filtered list of emitters whose name matched the
// driftSandboxEmitterPrefix. They're stopped and deleted unconditionally;
// when the schedule is non-empty a fresh emitter replaces them.
func (s *Signal) reconcileDriftSandboxEmitter(
	ctx workflow.Context,
	l log.Logger,
	nativeOn bool,
	install *app.Install,
	appCfg *app.AppConfig,
	queue *app.Queue,
	existing []app.QueueEmitter,
) error {
	stopAndDeleteEmitters(ctx, l, nativeOn, existing)

	schedule := appCfg.SandboxConfig.DriftSchedule
	if schedule == "" {
		return nil
	}

	name := driftSandboxEmitterPrefix + install.ID
	em, err := emitterclient.AwaitCreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:      queue.ID,
		Name:         name,
		Description:  fmt.Sprintf("sandbox drift check for install %s", install.ID),
		Mode:         app.QueueEmitterModeCron,
		CronSchedule: schedule,
		SignalType:   driftchecksandbox.SignalType,
		SignalTemplate: &driftchecksandbox.Signal{
			InstallID: install.ID,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create sandbox drift emitter %s: %w", name, err)
	}

	if nativeOn {
		if err := emitterclient.AwaitEnsureSchedule(ctx, em.ID); err != nil {
			return fmt.Errorf("unable to ensure sandbox drift schedule %s: %w", name, err)
		}
	}

	return nil
}
