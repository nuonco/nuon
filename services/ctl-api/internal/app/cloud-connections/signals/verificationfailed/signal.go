package verificationfailed

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "cloud-connection-verification-failed"

type Signal struct {
	ConnectionID   string            `json:"connection_id"`
	ConnectionName string            `json:"connection_name"`
	OrgID          string            `json:"org_id"`
	Platform       app.CloudPlatform `json:"platform"`
	Message        string            `json:"message"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType        { return SignalType }
func (s *Signal) Execute(workflow.Context) error { return nil }

func (s *Signal) Validate(workflow.Context) error {
	if s.ConnectionID == "" || s.OrgID == "" || s.Platform == "" || s.Message == "" {
		return fmt.Errorf("cloud connection verification failure payload is incomplete")
	}
	return nil
}

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		OrgID: s.OrgID, Operation: "verification_failed", OwnerID: s.ConnectionID,
		OwnerType: "cloud_connections", OwnerName: s.ConnectionName,
		Metadata: map[string]any{"connection_id": s.ConnectionID, "connection_name": s.ConnectionName, "platform": string(s.Platform), "message": s.Message},
	}
}
