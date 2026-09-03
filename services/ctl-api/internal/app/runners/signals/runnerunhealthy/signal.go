package runnerunhealthy

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "runner-unhealthy"

type Signal struct {
	RunnerID             string              `json:"runner_id"`
	RunnerName           string              `json:"runner_name,omitempty"`
	OrgID                string              `json:"org_id"`
	OrgName              string              `json:"org_name,omitempty"`
	FromStatus           app.RunnerStatus    `json:"from_status"`
	ToStatus             app.RunnerStatus    `json:"to_status"`
	Reason               string              `json:"reason"`
	RunnerGroupID        string              `json:"runner_group_id"`
	RunnerGroupType      app.RunnerGroupType `json:"runner_group_type"`
	RunnerGroupOwnerID   string              `json:"runner_group_owner_id"`
	RunnerGroupOwnerType string              `json:"runner_group_owner_type"`
	RunnerGroupOwnerName string              `json:"runner_group_owner_name,omitempty"`
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithAutoRetry        = (*Signal)(nil)
	_ signal.SignalWithMaxRetries       = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType { return SignalType }
func (s *Signal) AutoRetry() bool         { return true }
func (s *Signal) MaxRetries() int         { return 5 }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	var installID *string
	if s.RunnerGroupOwnerType == "installs" {
		installID = &s.RunnerGroupOwnerID
	}

	return signal.SignalLifecycleContext{
		OrgID:     s.OrgID,
		OrgName:   s.OrgName,
		InstallID: installID,
		Operation: "unhealthy",
		OwnerID:   s.RunnerGroupOwnerID,
		OwnerType: s.RunnerGroupOwnerType,
		OwnerName: s.RunnerGroupOwnerName,
		Metadata: map[string]any{
			"runner_id":               s.RunnerID,
			"runner_name":             s.RunnerName,
			"from_status":             string(s.FromStatus),
			"to_status":               string(s.ToStatus),
			"reason":                  s.Reason,
			"runner_group_id":         s.RunnerGroupID,
			"runner_group_type":       string(s.RunnerGroupType),
			"runner_group_owner_id":   s.RunnerGroupOwnerID,
			"runner_group_owner_type": s.RunnerGroupOwnerType,
		},
	}
}

func (s *Signal) Validate(_ workflow.Context) error {
	if s.RunnerID == "" || s.OrgID == "" || s.Reason == "" || s.RunnerGroupID == "" ||
		s.RunnerGroupType == "" || s.RunnerGroupOwnerID == "" || s.RunnerGroupOwnerType == "" {
		return errors.New("runner unhealthy payload is incomplete")
	}
	if s.ToStatus != app.RunnerStatusOffline || s.FromStatus == app.RunnerStatusOffline {
		return errors.New("runner unhealthy requires a transition to offline")
	}
	return nil
}

func (s *Signal) Execute(_ workflow.Context) error { return nil }
