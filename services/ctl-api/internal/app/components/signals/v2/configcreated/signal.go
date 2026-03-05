package configcreated

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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
	// Same logic as queuebuild - get component then queue a build
	cmp, err := activities.AwaitGetComponentByComponentID(ctx, s.ComponentID)
	if err != nil {
		return fmt.Errorf("unable to get component: %w", err)
	}

	_, err = activities.AwaitQueueComponentBuild(ctx, activities.QueueComponentBuildRequest{
		CreatedByID: cmp.CreatedByID,
		ComponentID: s.ComponentID,
		OrgID:       cmp.OrgID,
	})
	if err != nil {
		return fmt.Errorf("unable to queue component build: %w", err)
	}

	return nil
}
