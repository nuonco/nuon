package appconfigupdated

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/executeactionworkflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

func (s *Signal) reconcileActionCronEmitters(
	ctx workflow.Context,
	l log.Logger,
	install *app.Install,
	appCfg *app.AppConfig,
	queue *app.Queue,
	existing []app.QueueEmitter,
) error {
	// Stop and delete all existing action cron emitters
	stopAndDeleteEmitters(ctx, l, existing)

	// Fetch install action workflows
	actionWorkflows, err := activities.AwaitGetActionWorkflows(ctx, &activities.GetActionWorkflows{
		InstallID: s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to get action workflows: %w", err)
	}

	iawByActionWorkflowID := make(map[string]*app.InstallActionWorkflow, len(actionWorkflows))
	for _, iaw := range actionWorkflows {
		iawByActionWorkflowID[iaw.ActionWorkflowID] = iaw
	}

	// Create emitters for each action workflow with a cron trigger
	for _, awc := range appCfg.ActionWorkflowConfigs {
		if awc.CronTrigger == nil {
			continue
		}

		iaw, ok := iawByActionWorkflowID[awc.ActionWorkflowID]
		if !ok {
			continue
		}

		name := actionCronEmitterPrefix + iaw.ID
		if _, err := emitterclient.AwaitCreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
			QueueID:         queue.ID,
			Name:            name,
			Description:     fmt.Sprintf("action cron for install %s, action workflow %s", s.InstallID, iaw.ActionWorkflowID),
			Mode:            app.QueueEmitterModeCron,
			CronSchedule:    awc.CronTrigger.CronSchedule,
			SignalExpiresIn: 15 * time.Minute,
			SignalType:      executeactionworkflow.SignalType,
			SignalTemplate: &executeactionworkflow.Signal{
				Signal: &actionworkflowrun.Signal{
					InstallID:               install.ID,
					InstallActionWorkflowID: iaw.ID,
					TriggerType:             app.ActionWorkflowTriggerTypeCron,
					TriggeredByType:         "cron",
					RunEnvVars:              map[string]string{"TRIGGER_TYPE": "cron"},
				},
			},
		}); err != nil {
			return fmt.Errorf("unable to create action cron emitter %s: %w", name, err)
		}
	}

	return nil
}
