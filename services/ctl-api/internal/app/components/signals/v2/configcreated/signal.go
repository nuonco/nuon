package configcreated

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	buildsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/build"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	orgprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/provision"
	orgreprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/reprovision"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

type Signal struct {
	ComponentID string `json:"component_id" validate:"required"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.ComponentID == "" {
		return errors.New("component_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	cmp, err := activities.AwaitGetComponentByComponentID(ctx, s.ComponentID)
	if err != nil {
		return fmt.Errorf("unable to get component: %w", err)
	}

	// Ensure org is provisioned before creating a build.
	if err := queueclient.EnsureQueueSignal(ctx, cmp.OrgID, "orgs", orgprovision.SignalType, orgreprovision.SignalType); err != nil {
		return fmt.Errorf("org provision not ready: %w", err)
	}

	build, err := activities.AwaitCreateComponentBuildRecord(ctx, activities.CreateComponentBuildRecordRequest{
		CreatedByID: cmp.CreatedByID,
		ComponentID: s.ComponentID,
		OrgID:       cmp.OrgID,
	})
	if err != nil {
		return fmt.Errorf("unable to queue component build: %w", err)
	}

	// Enqueue the build signal to the component's queue. Fire and forget —
	// the build signal runs independently on the component queue.
	_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   s.ComponentID,
		OwnerType: "components",
		Signal: &buildsignal.Signal{
			ComponentID: s.ComponentID,
			BuildID:     build.ID,
		},
		SignalOwnerID:   build.ID,
		SignalOwnerType: "component_builds",
	})
	if err != nil {
		return fmt.Errorf("unable to enqueue build signal: %w", err)
	}

	return nil
}
