package terminateworkflows

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	orgactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "org-terminate-workflows"

type Signal struct {
	OrgID string `json:"org_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	_, err := orgactivities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("org not found: %w", err)
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)
	l.Info("terminating all workflows for org", zap.String("org_id", s.OrgID))

	resp, err := activities.AwaitTerminateOrgWorkflowsByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to terminate org workflows: %w", err)
	}

	l.Info("org workflow termination complete",
		zap.Int("terminated", resp.Terminated),
		zap.Int("errors", len(resp.Errors)))

	return nil
}
