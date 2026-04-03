package restart

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	actionssignals "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/signals"
	appssignals "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
	componentssignals "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	installssignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	runnerssignals "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signalsactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/signals/activities"
)

const SignalType signal.SignalType = "org-restart"

type Signal struct {
	OrgID string `json:"org_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	_, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("org not found: %w", err)
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	evs, err := activities.AwaitGetEventLoopsByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get event loops: %w", err)
	}

	for _, ev := range evs {
		if ev.WorkflowRef != nil {
			queueclient.AwaitRestart(ctx, ev.ID)
			continue
		}

		if err := s.restartEventLoop(ctx, l, ev.Namespace, ev.ID); err != nil {
			return fmt.Errorf("unable to restart event loop: %w", err)
		}
	}

	return nil
}

func (s *Signal) restartEventLoop(ctx workflow.Context, l log.Logger, namespace, id string) error {
	switch namespace {
	case "orgs":
		// skip restarting own namespace to avoid infinite loop
		return nil
	case "apps":
		signalsactivities.AwaitPkgSignalsSendAppsSignal(ctx, &signalsactivities.SendSignalRequest[*appssignals.Signal]{
			ID:     id,
			Signal: &appssignals.Signal{Type: appssignals.OperationRestart},
		})
	case "components":
		signalsactivities.AwaitPkgSignalsSendComponentsSignal(ctx, &signalsactivities.SendSignalRequest[*componentssignals.Signal]{
			ID:     id,
			Signal: &componentssignals.Signal{Type: componentssignals.OperationRestart},
		})
	case "runners":
		signalsactivities.AwaitPkgSignalsSendRunnersSignal(ctx, &signalsactivities.SendSignalRequest[*runnerssignals.Signal]{
			ID:     id,
			Signal: &runnerssignals.Signal{Type: runnerssignals.OperationRestart},
		})
	case "installs":
		signalsactivities.AwaitPkgSignalsSendInstallsSignal(ctx, &signalsactivities.SendSignalRequest[*installssignals.Signal]{
			ID:     id,
			Signal: &installssignals.Signal{Type: installssignals.OperationRestart},
		})
	case "actions":
		signalsactivities.AwaitPkgSignalsSendActionsSignal(ctx, &signalsactivities.SendSignalRequest[*actionssignals.Signal]{
			ID:     id,
			Signal: &actionssignals.Signal{Type: actionssignals.OperationRestart},
		})
	default:
		l.Error("unhandled namespace for restart", zap.String("namespace", namespace))
	}

	return nil
}
