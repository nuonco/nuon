package appconfigupdated

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/registryprune"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

const registryPruneSchedule = "0 9 * * *"

func (s *Signal) reconcileRegistryPruneEmitter(
	ctx workflow.Context,
	l log.Logger,
	install *app.Install,
	queue *app.Queue,
	existing []app.QueueEmitter,
) error {
	stopAndDeleteEmitters(ctx, l, existing)

	name := registryPruneEmitterPrefix + install.ID
	if _, err := emitterclient.AwaitCreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:        queue.ID,
		Name:           name,
		Description:    fmt.Sprintf("registry prune for install %s", install.ID),
		Mode:           app.QueueEmitterModeCron,
		CronSchedule:   registryPruneSchedule,
		JitterWindow:   driftJitterWindow,
		SignalType:     registryprune.SignalType,
		SignalTemplate: &registryprune.Signal{InstallID: install.ID},
	}); err != nil {
		return fmt.Errorf("unable to create registry prune emitter %s: %w", name, err)
	}
	return nil
}
