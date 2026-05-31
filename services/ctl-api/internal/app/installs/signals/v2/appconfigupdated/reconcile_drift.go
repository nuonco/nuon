package appconfigupdated

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/driftcheck"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

func (s *Signal) reconcileDriftEmitters(
	ctx workflow.Context,
	l log.Logger,
	install *app.Install,
	appCfg *app.AppConfig,
	queue *app.Queue,
	existing []app.QueueEmitter,
) error {
	// Stop and delete all existing drift emitters
	stopAndDeleteEmitters(ctx, l, existing)

	// Fetch install components
	installComponents, err := activities.AwaitGetInstallComponentsByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install components: %w", err)
	}

	icByComponentID := make(map[string]app.InstallComponent, len(installComponents))
	for _, ic := range installComponents {
		icByComponentID[ic.ComponentID] = ic
	}

	// Create emitters for each component with a drift schedule
	for _, ccc := range appCfg.ComponentConfigConnections {
		if ccc.DriftSchedule == "" {
			continue
		}

		ic, ok := icByComponentID[ccc.ComponentID]
		if !ok {
			continue
		}

		name := driftEmitterPrefix + ic.ID
		if _, err := emitterclient.AwaitCreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
			QueueID:         queue.ID,
			Name:            name,
			Description:     fmt.Sprintf("drift check for install %s, component %s", s.InstallID, ic.ComponentID),
			Mode:            app.QueueEmitterModeCron,
			CronSchedule:    ccc.DriftSchedule,
			SignalExpiresIn: 15 * time.Minute,
			SignalType:      driftcheck.SignalType,
			SignalTemplate: &driftcheck.Signal{
				InstallID:          install.ID,
				InstallComponentID: ic.ID,
				ComponentID:        ic.ComponentID,
			},
		}); err != nil {
			return fmt.Errorf("unable to create drift emitter %s: %w", name, err)
		}
	}

	return nil
}
