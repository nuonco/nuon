package temporal

import (
	"errors"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"

	actionssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/signals"
	appssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	componentssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	generalsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
	installssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	orgssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	releasessignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/signals"
	runnerssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	signalsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/signals/activities"
)

const (
	defaultSignalSendTimeout time.Duration = time.Second * 5
)

func (e *evClient) Send(ctx workflow.Context, id string, signal eventloop.Signal) {
	var err error
	switch signal.Namespace() {
	case actionssignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendActionsSignal(ctx, &signalsactivities.SendSignalRequest[*actionssignals.Signal]{
			ID:     id,
			Signal: signal.(*actionssignals.Signal),
		})
	case appssignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendAppsSignal(ctx, &signalsactivities.SendSignalRequest[*appssignals.Signal]{
			ID:     id,
			Signal: signal.(*appssignals.Signal),
		})
	case installssignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendInstallsSignal(ctx, &signalsactivities.SendSignalRequest[*installssignals.Signal]{
			ID:     id,
			Signal: signal.(*installssignals.Signal),
		})
	case componentssignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendComponentsSignal(ctx, &signalsactivities.SendSignalRequest[*componentssignals.Signal]{
			ID:     id,
			Signal: signal.(*componentssignals.Signal),
		})
	case orgssignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendOrgsSignal(ctx, &signalsactivities.SendSignalRequest[*orgssignals.Signal]{
			ID:     id,
			Signal: signal.(*orgssignals.Signal),
		})
	case releasessignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendReleasesSignal(ctx, &signalsactivities.SendSignalRequest[*releasessignals.Signal]{
			ID:     id,
			Signal: signal.(*releasessignals.Signal),
		})
	case runnerssignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendRunnersSignal(ctx, &signalsactivities.SendSignalRequest[*runnerssignals.Signal]{
			ID:     id,
			Signal: signal.(*runnerssignals.Signal),
		})
	case generalsignals.TemporalNamespace:
		signalsactivities.AwaitPkgSignalsSendGeneralSignal(ctx, &signalsactivities.SendSignalRequest[*generalsignals.Signal]{
			ID:     id,
			Signal: signal.(*generalsignals.Signal),
		})
	default:
		err = errors.New("unsupported namespace " + signal.Namespace())
	}

	if err != nil {
		e.l.Error("unable to send signal",
			zap.String("id", id),
			zap.String("namespace", signal.Namespace()),
			zap.Error(err),
		)
	}
}

func (e *evClient) Cancel(wCtx workflow.Context, namespace, id string) {
	//
}
