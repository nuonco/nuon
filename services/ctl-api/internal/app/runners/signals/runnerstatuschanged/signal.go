package runnerstatuschanged

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "runner_status_changed"

type Signal struct {
	RunnerID        string              `json:"runner_id"`
	OrgID           string              `json:"org_id"`
	FromStatus      app.RunnerStatus    `json:"from_status"`
	ToStatus        app.RunnerStatus    `json:"to_status"`
	Reason          string              `json:"reason"`
	RunnerGroupType app.RunnerGroupType `json:"runner_group_type"`
	OwnerID         string              `json:"owner_id"`
	OwnerType       string              `json:"owner_type"`
}

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(workflow.Context) error {
	if s.RunnerID == "" || s.OrgID == "" || s.FromStatus == "" || s.ToStatus == "" || s.RunnerGroupType == "" || s.OwnerID == "" || s.OwnerType == "" {
		return errors.New("runner status changed payload is incomplete")
	}
	return nil
}

func (s *Signal) Execute(workflow.Context) error { return nil }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	ctx := signal.SignalLifecycleContext{
		OrgID: s.OrgID, Operation: string(s.ToStatus), OwnerID: s.RunnerID, OwnerType: "runners",
		Metadata: map[string]any{
			"runner_id": s.RunnerID, "from_status": string(s.FromStatus), "to_status": string(s.ToStatus), "reason": s.Reason,
			"runner_group_type": string(s.RunnerGroupType), "group_owner_id": s.OwnerID, "group_owner_type": s.OwnerType,
		},
	}
	if s.OwnerType == "installs" {
		ctx.InstallID = &s.OwnerID
	}
	return ctx
}

var _ signal.SignalWithLifecycleContext = (*Signal)(nil)
