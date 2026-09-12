package handler

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/sandboxmode"
)

// todo(sk): clean up after terminating old workflows
//
// sandboxModeCacheVersion gates caching the org's sandbox-mode flag across the
// validate and execute phases. Histories written before the cache recorded one
// GetOrgByID per phase, so replaying them against a single fetch is
// nondeterministic.
const sandboxModeCacheVersion = "queue-handler-sandbox-mode-cache-v1"

func (h *handler) checkSandboxMode(ctx workflow.Context) (signal.Signal, error) {
	if h.queueSignal.OrgID == nil {
		return h.sig, nil
	}

	cacheOK := workflow.GetVersion(ctx, sandboxModeCacheVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion

	if !cacheOK || !h.sandboxModeResolved {
		org, err := activities.AwaitGetOrgByIDByOrgID(ctx, generics.FromPtrStr(h.queueSignal.OrgID))
		if err != nil {
			return nil, errors.Wrap(err, "unable to get org")
		}
		h.sandboxMode = org.SandboxMode
		h.sandboxModeResolved = true
	}

	if !h.sandboxMode {
		return h.sig, nil
	}

	return sandboxmode.WrapSignal(h.sig), nil
}
