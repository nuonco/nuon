package appconfigupdate

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app_config_updated"

type Signal struct {
	InstallID string `json:"install_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// 1. Fetch install
	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	// 2. Fetch full app config
	appConfig, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}
	if appConfig == nil || appConfig.ID == "" {
		return nil
	}

	// 3. Find the install-signals queue
	var queue *app.Queue
	err = workflow.ExecuteActivity(ctx, queueclient.AwaitGetQueueByOwnerAndName, s.InstallID, "installs", "install-signals").Get(ctx, &queue)
	if err != nil {
		return errors.Wrap(err, "unable to get install-signals queue")
	}

	// 4. Get all existing emitters for this queue
	existingEmitters, err := emitterclient.AwaitGetEmittersByQueueID(ctx, queue.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get existing emitters")
	}
	emittersByName := make(map[string]app.QueueEmitter, len(existingEmitters))
	for _, em := range existingEmitters {
		emittersByName[em.Name] = em
	}

	// 5. Reconcile action workflow cron emitters
	if err := s.reconcileActionCronEmitters(ctx, install, appConfig, queue.ID, emittersByName); err != nil {
		return err
	}

	// 6. Reconcile component drift emitters
	if err := s.reconcileComponentDriftEmitters(ctx, install, appConfig, queue.ID, emittersByName); err != nil {
		return err
	}

	// 7. Reconcile sandbox drift emitter
	if err := s.reconcileSandboxDriftEmitter(ctx, appConfig, queue.ID, emittersByName); err != nil {
		return err
	}

	return nil
}

func (s *Signal) reconcileActionCronEmitters(
	ctx workflow.Context,
	install *app.Install,
	appConfig *app.AppConfig,
	queueID string,
	existing map[string]app.QueueEmitter,
) error {
	// Fetch install action workflows
	iaws, err := activities.AwaitGetActionWorkflowsByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install action workflows")
	}

	iawByAWID := make(map[string]*app.InstallActionWorkflow, len(iaws))
	for _, iaw := range iaws {
		iawByAWID[iaw.ActionWorkflowID] = iaw
	}

	for _, awc := range appConfig.ActionWorkflowConfigs {
		iaw, ok := iawByAWID[awc.ActionWorkflowID]
		if !ok {
			continue
		}

		emitterName := fmt.Sprintf("action-cron-%s", iaw.ID)
		em, exists := existing[emitterName]

		var cronSchedule string
		if awc.CronTrigger != nil {
			cronSchedule = awc.CronTrigger.CronSchedule
		}

		req := &emitterclient.ReconcileActionCronEmitterRequest{
			InstallID:               s.InstallID,
			QueueID:                 queueID,
			InstallActionWorkflowID: iaw.ID,
			ActionWorkflowID:        awc.ActionWorkflowID,
			CronSchedule:            cronSchedule,
		}
		if exists {
			req.ExistingEmitterID = em.ID
			req.ExistingCronSchedule = em.CronSchedule
		}

		if err := emitterclient.AwaitReconcileActionCronEmitter(ctx, req); err != nil {
			return errors.Wrapf(err, "unable to reconcile action cron emitter for %s", iaw.ID)
		}
	}

	return nil
}

func (s *Signal) reconcileComponentDriftEmitters(
	ctx workflow.Context,
	install *app.Install,
	appConfig *app.AppConfig,
	queueID string,
	existing map[string]app.QueueEmitter,
) error {
	for _, ccc := range appConfig.ComponentConfigConnections {
		// Fetch the install component for this component
		ic, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
			InstallID:   s.InstallID,
			ComponentID: ccc.ComponentID,
		})
		if err != nil {
			return errors.Wrapf(err, "unable to get install component for %s", ccc.ComponentID)
		}
		if ic == nil {
			continue
		}

		// Skip component types that don't support drift
		switch ic.Component.Type {
		case app.ComponentTypeDockerBuild, app.ComponentTypeExternalImage:
			continue
		}

		emitterName := fmt.Sprintf("drift-component-%s", ic.ID)
		em, exists := existing[emitterName]

		req := &emitterclient.ReconcileComponentDriftEmitterRequest{
			InstallID:          s.InstallID,
			QueueID:            queueID,
			InstallComponentID: ic.ID,
			ComponentID:        ccc.ComponentID,
			ComponentName:      ic.Component.Name,
			DriftSchedule:      ccc.DriftSchedule,
		}
		if exists {
			req.ExistingEmitterID = em.ID
			req.ExistingSchedule = em.CronSchedule
		}

		// Get the latest build if drift is enabled
		if ccc.DriftSchedule != "" {
			build, err := activities.AwaitGetComponentLatestBuildByComponentID(ctx, ccc.ComponentID)
			if err != nil {
				// No build yet, skip this component
				continue
			}
			req.ComponentBuildID = build.ID
		}

		if err := emitterclient.AwaitReconcileComponentDriftEmitter(ctx, req); err != nil {
			return errors.Wrapf(err, "unable to reconcile component drift emitter for %s", ccc.ComponentID)
		}
	}

	return nil
}

func (s *Signal) reconcileSandboxDriftEmitter(
	ctx workflow.Context,
	appConfig *app.AppConfig,
	queueID string,
	existing map[string]app.QueueEmitter,
) error {
	emitterName := fmt.Sprintf("drift-sandbox-%s", s.InstallID)
	em, exists := existing[emitterName]

	req := &emitterclient.ReconcileSandboxDriftEmitterRequest{
		InstallID:     s.InstallID,
		QueueID:       queueID,
		DriftSchedule: appConfig.SandboxConfig.DriftSchedule,
	}
	if exists {
		req.ExistingEmitterID = em.ID
		req.ExistingSchedule = em.CronSchedule
	}

	return emitterclient.AwaitReconcileSandboxDriftEmitter(ctx, req)
}
