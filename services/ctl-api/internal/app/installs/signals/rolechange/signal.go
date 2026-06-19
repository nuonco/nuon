package rolechange

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/actions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "role-change"

const installSignalsQueueName = "install-signals"

type Signal struct {
	InstallID      string `json:"install_id"`
	RoleName       string `json:"role_name"`
	RoleType       string `json:"role_type"`
	ChangeType     string `json:"change_type"`
	RoleID         string `json:"role_id"`
	InstallRolesID string `json:"install_roles_id"`
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
	installID := &s.InstallID
	if s.InstallID == "" {
		installID = nil
	}
	return signal.SignalLifecycleContext{
		InstallID: installID,
		Operation: "role-change",
		OwnerID:   s.InstallID,
		OwnerType: "installs",
		Metadata: map[string]any{
			"role_name":   s.RoleName,
			"role_type":   s.RoleType,
			"change_type": s.ChangeType,
			"role_id":     s.RoleID,
		},
	}
}

func (s *Signal) Validate(_ workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	if s.RoleName == "" {
		return errors.New("role_name is required")
	}
	if s.ChangeType == "" {
		return errors.New("change_type is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	triggerType := s.triggerType()

	runEnvVars := map[string]*string{
		"TRIGGER_TYPE": strPtr(string(triggerType)),
		"ROLE_NAME":    strPtr(s.RoleName),
		"ROLE_TYPE":    strPtr(s.RoleType),
		"CHANGE_TYPE":  strPtr(s.ChangeType),
		"ROLE_ID":      strPtr(s.RoleID),
		"ROLE_ARN":     strPtr(s.RoleID),
	}

	wfID := fmt.Sprintf("role-change-actions-%s-%s-%s", s.InstallID, s.ChangeType, s.RoleName)
	if err := actions.AwaitLifecycleActionWorkflows(ctx, &actions.LifecycleActionWorkflowsRequest{
		InstallID:   s.InstallID,
		TriggerType: triggerType,
		RunEnvVars:  runEnvVars,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: wfID,
	}); err != nil {
		l.Warn("unable to execute role-change action workflows",
			zap.String("install_id", s.InstallID),
			zap.String("change_type", s.ChangeType),
			zap.Error(err))
	}

	return nil
}

func (s *Signal) triggerType() app.ActionWorkflowTriggerType {
	if s.ChangeType == "enabled" {
		return app.ActionWorkflowTriggerTypeRoleEnabled
	}
	return app.ActionWorkflowTriggerTypeRoleDisabled
}

func strPtr(s string) *string {
	return &s
}
