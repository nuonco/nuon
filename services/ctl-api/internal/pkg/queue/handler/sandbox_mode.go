package handler

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/sandboxmode"
)

func (h *handler) checkSandboxMode(ctx workflow.Context) (signal.Signal, error) {
	if h.queueSignal.OrgID == nil {
		return h.sig, nil
	}

	org, err := activities.AwaitGetOrgByIDByOrgID(ctx, generics.FromPtrStr(h.queueSignal.OrgID))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org")
	}

	if !org.SandboxMode {
		return h.sig, nil
	}

	return sandboxmode.WrapSignal(h.sig), nil
}
