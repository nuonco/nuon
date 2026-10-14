package appconfigupdated

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/runbookcron"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

func (s *Signal) reconcileRunbookCronEmitters(
	ctx workflow.Context,
	l log.Logger,
	install *app.Install,
	queue *app.Queue,
	existing []app.QueueEmitter,
) error {
	stopAndDeleteEmitters(ctx, l, existing)

	crons, err := activities.AwaitGetInstallRunbookCrons(ctx, activities.GetInstallRunbookCronsRequest{
		InstallID:   install.ID,
		AppConfigID: install.AppConfigID,
	})
	if err != nil {
		return fmt.Errorf("unable to get runbook crons: %w", err)
	}

	for _, rc := range crons {
		name := runbookCronEmitterPrefix + rc.InstallRunbookID
		if _, err := emitterclient.AwaitCreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
			QueueID:         queue.ID,
			Name:            name,
			Description:     fmt.Sprintf("runbook cron for install %s, install runbook %s", install.ID, rc.InstallRunbookID),
			Mode:            app.QueueEmitterModeCron,
			CronSchedule:    rc.CronSchedule,
			JitterWindow:    actionCronJitterWindow,
			SignalExpiresIn: 15 * time.Minute,
			SignalType:      runbookcron.SignalType,
			SignalTemplate: &runbookcron.Signal{
				InstallID:        install.ID,
				InstallRunbookID: rc.InstallRunbookID,
				RunbookConfigID:  rc.RunbookConfigID,
				TriggeredByID:    rc.TriggeredByID,
			},
		}); err != nil {
			return fmt.Errorf("unable to create runbook cron emitter %s: %w", name, err)
		}
	}

	return nil
}
