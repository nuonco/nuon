package operationroles

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestDefaultRoleForWorkflowType(t *testing.T) {
	appCfg := &app.AppConfig{
		PermissionsConfig: app.AppPermissionsConfig{
			ProvisionRole:   app.AppAWSIAMRoleConfig{Name: "provision"},
			DeprovisionRole: app.AppAWSIAMRoleConfig{Name: "deprovision"},
			MaintenanceRole: app.AppAWSIAMRoleConfig{Name: "maintenance"},
		},
	}

	tests := []struct {
		workflowType app.WorkflowType
		expected     string
	}{
		{app.WorkflowTypeProvision, "provision"},
		{app.WorkflowTypeReprovision, "provision"},
		{app.WorkflowTypeReprovisionSandbox, "provision"},
		{app.WorkflowTypeDriftRunReprovisionSandbox, "provision"},
		{app.WorkflowTypeDeprovision, "deprovision"},
		{app.WorkflowTypeDeprovisionSandbox, "deprovision"},
		{app.WorkflowTypeManualDeploy, "maintenance"},
		{app.WorkflowTypeInputUpdate, "maintenance"},
		{app.WorkflowTypeDeployComponents, "maintenance"},
		{app.WorkflowTypeTeardownComponent, "maintenance"},
		{app.WorkflowTypeTeardownComponents, "maintenance"},
		{app.WorkflowTypeActionWorkflowRun, "maintenance"},
		{app.WorkflowTypeSyncSecrets, "maintenance"},
		{app.WorkflowTypeDriftRun, "maintenance"},
		{app.WorkflowTypeRunbookRun, "maintenance"},
		{"", "maintenance"},
	}

	for _, tt := range tests {
		t.Run(string(tt.workflowType), func(t *testing.T) {
			assert.Equal(t, tt.expected, DefaultRoleForWorkflowType(appCfg, tt.workflowType))
		})
	}
}
