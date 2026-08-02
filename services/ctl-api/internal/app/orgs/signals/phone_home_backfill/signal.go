// Package phone_home_backfill onboards one org's installs to phone-home auth.
//
// It is step one of two. This backfills every install in the org — cloud metadata
// first, then the phone-home secret — while phone-home-auth is still off for the org.
// Step two is enabling the flag, which is the existing
// PATCH /v1/orgs/{org_id}/admin-features. Provisioning first means turning the flag on
// never races the credentials.
//
// The org is the organizing unit for this feature: the flag is a deliberate per-org
// admin action, target_account_id becomes required per org, and an operator onboards
// one org at a time. So the backfill is per org too, rather than a fleet-wide job whose
// aggregate counters say nothing about the org you care about.
//
// It fans out rather than looping because it cannot loop: this signal runs in the orgs
// namespace and both backfill activities are registered on the installs worker.
// EnqueueSignalToOwner is registered in every namespace and crosses the boundary by
// writing to each install's own queue — the same route restart_runners takes to reach
// runners.
package phone_home_backfill

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/phonehomebackfill"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "org-phone-home-backfill"

type Signal struct {
	OrgID string `json:"org_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	if _, err := activities.AwaitGetByOrgID(ctx, s.OrgID); err != nil {
		return fmt.Errorf("org not found: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	installs, err := activities.AwaitGetInstallsByID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get installs: %w", err)
	}

	for _, install := range installs {
		_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   install.ID,
			OwnerType: "installs",
			// Named explicitly: an install owns several queues, so the
			// owner-only lookup restart_runners relies on is ambiguous here.
			QueueName:      installhelpers.InstallSignalsQueueName,
			IdempotencyKey: fmt.Sprintf("phone-home-backfill-%s", install.ID),
			Signal: &phonehomebackfill.Signal{
				InstallID: install.ID,
				// The point of the backfill: metadata and secrets land while the
				// flag is still off, so turning it on never races provisioning.
				IgnoreOrgFeatureGate: true,
			},
		})
		if err != nil {
			return fmt.Errorf("unable to enqueue phone home secret signal for install %s: %w", install.ID, err)
		}
	}

	return nil
}
