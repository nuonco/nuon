package rolechange

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/executeactionworkflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "role-change"

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

// Execute enqueues an actionworkflowrun signal for each install action with a
// matching role-enabled/role-disabled trigger.
func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	triggerType := s.triggerType()

	runEnvVars := map[string]string{
		"TRIGGER_TYPE": string(triggerType),
		"ROLE_NAME":    s.RoleName,
		"ROLE_TYPE":    s.RoleType,
		"CHANGE_TYPE":  s.ChangeType,
		"ROLE_ID":      s.RoleID,
		"ROLE_ARN":     s.RoleID,
	}

	installActions, err := activities.AwaitGetActionWorkflowsByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install action workflows")
	}

	for _, installAction := range installActions {
		cfg, err := activities.AwaitGetActionWorkflowLatestConfigByActionWorkflowID(ctx, installAction.ActionWorkflowID)
		if err != nil {
			return errors.Wrap(err, "unable to get action workflow config")
		}
		if !hasLifecycleTrigger(cfg, triggerType) {
			continue
		}

		if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   s.InstallID,
			OwnerType: "installs",
			QueueName: queuenames.InstallActionWorkflowsQueueName,
			Signal: &executeactionworkflow.Signal{
				Signal: &actionworkflowrun.Signal{
					InstallID:               s.InstallID,
					InstallActionWorkflowID: installAction.ID,
					TriggerType:             triggerType,
					TriggeredByType:         string(triggerType),
					RunEnvVars:              runEnvVars,
				},
			},
		}); err != nil {
			l.Warn("unable to enqueue role-change action run",
				zap.String("install_id", s.InstallID),
				zap.String("install_action_workflow_id", installAction.ID),
				zap.Error(err))
		}
	}

	return nil
}

func hasLifecycleTrigger(cfg *app.ActionWorkflowConfig, triggerType app.ActionWorkflowTriggerType) bool {
	for _, trigger := range cfg.LifecycleTriggers {
		if trigger.Type == triggerType {
			return true
		}
	}
	return false
}

func (s *Signal) triggerType() app.ActionWorkflowTriggerType {
	if s.ChangeType == "enabled" {
		return app.ActionWorkflowTriggerTypeRoleEnabled
	}
	return app.ActionWorkflowTriggerTypeRoleDisabled
}
