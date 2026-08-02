// Package phonehomebackfill onboards one install to phone-home auth.
//
// Two steps in order, because the second cannot work without the first: backfill the
// install's cloud metadata from the identifier its stack already reported, then
// reconcile its phone-home secret. An install created before target_account_id existed
// carries no target, and the reconciler skips any install without one — so running the
// secret step alone against a pre-existing fleet does nothing.
//
// Both steps live in one signal rather than two so the ordering holds per install. Two
// separate fan-outs would leave the operator judging when the first had drained, with
// no completion signal to wait on.
//
// It also exists so the org-scoped backfill can reach these activities at all: that
// signal runs in the orgs namespace and both activities are registered on the installs
// worker, so the org signal fans one of these onto each install's own signals queue
// rather than calling them directly. A signal per install additionally means each
// install is retried and observable on its own, and one bad install cannot stall the
// rest of the org.
package phonehomebackfill

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "install-phone-home-backfill"

type Signal struct {
	InstallID string `json:"install_id"`
	// IgnoreOrgFeatureGate provisions ahead of the org enabling phone-home-auth,
	// which is the whole point of the backfill: metadata and secrets land first, the
	// flag goes on last. See the field of the same name on the activity request for
	// what it does and does not waive.
	IgnoreOrgFeatureGate bool `json:"ignore_org_feature_gate,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID: &s.InstallID,
		Operation: "phone-home-backfill",
	}
}

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	if _, err := activities.AwaitBackfillInstallCloudMetadataByInstallID(ctx, s.InstallID); err != nil {
		return fmt.Errorf("unable to backfill install cloud metadata: %w", err)
	}

	// The full request rather than the ByField helper, which cannot carry the gate
	// override.
	if _, err := activities.AwaitEnsureInstallPhoneHomeSecret(ctx, &activities.EnsureInstallPhoneHomeSecretRequest{
		InstallID:            s.InstallID,
		IgnoreOrgFeatureGate: s.IgnoreOrgFeatureGate,
	}); err != nil {
		return fmt.Errorf("unable to ensure install phone home secret: %w", err)
	}

	return nil
}
