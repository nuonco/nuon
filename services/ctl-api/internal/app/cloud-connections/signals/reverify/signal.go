package reverify

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "cloud_connection_reverify"

type Signal struct {
	CloudConnectionID string `json:"cloud_connection_id"`
	OnDemand          bool   `json:"on_demand,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType       { return SignalType }
func (s *Signal) MaxInFlightAge() time.Duration { return 5 * time.Minute }

func (s *Signal) Validate(workflow.Context) error {
	if s.CloudConnectionID == "" {
		return fmt.Errorf("cloud_connection_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	return activities.AwaitReverify(ctx, activities.ReverifyRequest{CloudConnectionID: s.CloudConnectionID, OnDemand: s.OnDemand})
}
