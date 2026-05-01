package reconcileemitters

import (
	"fmt"
	"strings"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/driftcheck"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/executeactionworkflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "reconcile-emitters"

const (
	actionCronEmitterPrefix = "action-cron-"
	driftEmitterPrefix      = "drift-"
)

type Signal struct {
	InstallID string `json:"install_id"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID: &s.InstallID,
		Operation: "reconcile-emitters",
	}
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}

	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("install not found: %w", err)
	}

	return nil
}

// emitterSpec describes a desired emitter configuration.
type emitterSpec struct {
	name         string
	cronSchedule string
	signalType   signal.SignalType
	signal       signal.Signal
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// 1. Fetch install
	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	// 2. Fetch the install's signals queue
	queue, err := activities.AwaitGetInstallSignalsQueueByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install signals queue: %w", err)
	}

	// 3. Fetch existing emitters on the queue
	existingEmitters, err := emitterclient.AwaitGetEmittersByQueueID(ctx, queue.ID)
	if err != nil {
		return fmt.Errorf("unable to get existing emitters: %w", err)
	}

	// 4. Fetch full app config
	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	// 5. Fetch install action workflows
	actionWorkflows, err := activities.AwaitGetActionWorkflows(ctx, &activities.GetActionWorkflows{
		InstallID: s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to get action workflows: %w", err)
	}

	// 6. Fetch install components
	installComponents, err := activities.AwaitGetInstallComponentsByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install components: %w", err)
	}

	// 7. Build desired emitters map
	desired := s.buildDesiredEmitters(install, appCfg, actionWorkflows, installComponents)

	// 8. Build existing emitters map (only managed emitters)
	existing := make(map[string]app.QueueEmitter)
	for _, em := range existingEmitters {
		if strings.HasPrefix(em.Name, actionCronEmitterPrefix) || strings.HasPrefix(em.Name, driftEmitterPrefix) {
			existing[em.Name] = em
		}
	}

	// 9. Delete stale emitters (exist but no longer needed, or schedule changed)
	for name, em := range existing {
		spec, needed := desired[name]
		if !needed || spec.cronSchedule != em.CronSchedule {
			l.Info("reconcile-emitters: removing stale emitter",
				zap.String("emitter_name", name),
				zap.String("emitter_id", em.ID))

			if _, err := emitterclient.AwaitStopEmitter(ctx, em.ID); err != nil {
				l.Warn("reconcile-emitters: unable to stop emitter",
					zap.String("emitter_id", em.ID),
					zap.Error(err))
			}
			if err := emitterclient.AwaitDeleteEmitter(ctx, em.ID); err != nil {
				l.Warn("reconcile-emitters: unable to delete emitter",
					zap.String("emitter_id", em.ID),
					zap.Error(err))
			}

			// If schedule changed, we need to recreate
			if needed {
				if _, err := emitterclient.AwaitCreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
					QueueID:        queue.ID,
					Name:           spec.name,
					Description:    fmt.Sprintf("auto-managed emitter for install %s", s.InstallID),
					Mode:           app.QueueEmitterModeCron,
					CronSchedule:   spec.cronSchedule,
					SignalType:     spec.signalType,
					SignalTemplate: spec.signal,
				}); err != nil {
					return fmt.Errorf("unable to recreate emitter %s: %w", spec.name, err)
				}
			}
		}
	}

	// 10. Create missing emitters
	for name, spec := range desired {
		if _, exists := existing[name]; exists {
			// Already handled above (either kept or recreated)
			continue
		}

		l.Info("reconcile-emitters: creating emitter",
			zap.String("emitter_name", name))

		if _, err := emitterclient.AwaitCreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
			QueueID:        queue.ID,
			Name:           spec.name,
			Description:    fmt.Sprintf("auto-managed emitter for install %s", s.InstallID),
			Mode:           app.QueueEmitterModeCron,
			CronSchedule:   spec.cronSchedule,
			SignalType:     spec.signalType,
			SignalTemplate: spec.signal,
		}); err != nil {
			return fmt.Errorf("unable to create emitter %s: %w", spec.name, err)
		}
	}

	return nil
}

func (s *Signal) buildDesiredEmitters(
	install *app.Install,
	appCfg *app.AppConfig,
	actionWorkflows []*app.InstallActionWorkflow,
	installComponents []app.InstallComponent,
) map[string]emitterSpec {
	desired := make(map[string]emitterSpec)

	// Build a lookup from ActionWorkflowID -> InstallActionWorkflow
	iawByActionWorkflowID := make(map[string]*app.InstallActionWorkflow)
	for _, iaw := range actionWorkflows {
		iawByActionWorkflowID[iaw.ActionWorkflowID] = iaw
	}

	// Action cron emitters
	for _, awc := range appCfg.ActionWorkflowConfigs {
		if awc.CronTrigger == nil {
			continue
		}

		iaw, ok := iawByActionWorkflowID[awc.ActionWorkflowID]
		if !ok {
			continue
		}

		name := actionCronEmitterPrefix + iaw.ID
		desired[name] = emitterSpec{
			name:         name,
			cronSchedule: awc.CronTrigger.CronSchedule,
			signalType:   executeactionworkflow.SignalType,
			signal: &executeactionworkflow.Signal{
				Signal: &actionworkflowrun.Signal{
					InstallID:               install.ID,
					InstallActionWorkflowID: iaw.ID,
					TriggerType:             app.ActionWorkflowTriggerTypeCron,
					TriggeredByType:         "cron",
					RunEnvVars:              map[string]string{"TRIGGER_TYPE": "cron"},
				},
			},
		}
	}

	// Build a lookup from ComponentID -> InstallComponent
	icByComponentID := make(map[string]app.InstallComponent)
	for _, ic := range installComponents {
		icByComponentID[ic.ComponentID] = ic
	}

	// Drift emitters
	for _, ccc := range appCfg.ComponentConfigConnections {
		if ccc.DriftSchedule == "" {
			continue
		}

		ic, ok := icByComponentID[ccc.ComponentID]
		if !ok {
			continue
		}

		name := driftEmitterPrefix + ic.ID
		desired[name] = emitterSpec{
			name:         name,
			cronSchedule: ccc.DriftSchedule,
			signalType:   driftcheck.SignalType,
			signal: &driftcheck.Signal{
				InstallID:          install.ID,
				InstallComponentID: ic.ID,
				ComponentID:        ic.ComponentID,
			},
		}
	}

	return desired
}
