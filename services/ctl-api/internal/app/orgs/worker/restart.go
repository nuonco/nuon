package worker

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	appssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	componentssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	installssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	orgssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	runnerssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 10s
func (w *Workflows) Restart(ctx workflow.Context, sreq signals.RequestSignal) error {
	evs, err := activities.AwaitGetEventLoopsByOrgID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get event loops")
	}

	for _, ev := range evs {
		if err := w.restartEventLoop(ctx, ev.Namespace, ev.ID); err != nil {
			return errors.Wrap(err, "unable to restart event loop")
		}
		workflow.Sleep(ctx, time.Millisecond)
	}

	return nil
}

func (w *Workflows) restartEventLoop(ctx workflow.Context, namespace, id string) error {
	switch namespace {
	case "orgs":
		return nil
		w.ev.Send(ctx, id, &orgssignals.Signal{
			Type: orgssignals.OperationRestart,
		})
	case "apps":
		w.ev.Send(ctx, id, &appssignals.Signal{
			Type: appssignals.OperationRestart,
		})
	case "components":
		w.ev.Send(ctx, id, &componentssignals.Signal{
			Type: componentssignals.OperationRestart,
		})
	case "runners":
		w.ev.Send(ctx, id, &runnerssignals.Signal{
			Type: runnerssignals.OperationRestart,
		})
	case "installs":
		w.ev.Send(ctx, id, &installssignals.Signal{
			Type: installssignals.OperationRestart,
		})
	default:
		return errors.New("unhandled namespace " + namespace)
	}

	return nil
}
