package operationroles

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestGetRoleForActionRejectsUnknownRuntimeRole(t *testing.T) {
	appCfg := &app.AppConfig{
		PermissionsConfig: app.AppPermissionsConfig{
			MaintenanceRole: app.AppAWSIAMRoleConfig{Name: "install-maintenance"},
		},
	}
	run := &app.InstallActionWorkflowRun{
		Role: "provision",
		ActionWorkflowConfig: app.ActionWorkflowConfig{
			ActionWorkflow: app.ActionWorkflow{Name: "diagnostics"},
		},
	}
	stack := &app.InstallStack{
		InstallStackOutputs: app.InstallStackOutputs{
			AWSStackOutputs: &app.AWSStackOutputs{
				MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/install-maintenance",
			},
		},
	}

	selection, operation, err := GetRoleForAction(
		zap.NewNop(),
		appCfg,
		run,
		stack,
		&state.State{ID: "inst_123", Name: "example"},
		nil,
	)

	require.ErrorContains(t, err, `unable to use requested role "provision"`)
	require.Nil(t, selection)
	require.Empty(t, operation)
}
