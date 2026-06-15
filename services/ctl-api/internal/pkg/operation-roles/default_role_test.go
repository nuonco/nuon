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

func TestDefaultRoleForWorkflow(t *testing.T) {
	appCfg := &app.AppConfig{
		PermissionsConfig: app.AppPermissionsConfig{
			ProvisionRole:   app.AppAWSIAMRoleConfig{Name: "provision"},
			DeprovisionRole: app.AppAWSIAMRoleConfig{Name: "deprovision"},
			MaintenanceRole: app.AppAWSIAMRoleConfig{Name: "maintenance"},
		},
	}

	tests := []struct {
		name     string
		flw      *app.Workflow
		expected string
	}{
		{"nil workflow falls back to maintenance", nil, "maintenance"},
		{"provision workflow uses provision role", &app.Workflow{Type: app.WorkflowTypeProvision}, "provision"},
		{"deprovision workflow uses deprovision role", &app.Workflow{Type: app.WorkflowTypeDeprovision}, "deprovision"},
		{"manual deploy uses maintenance role", &app.Workflow{Type: app.WorkflowTypeManualDeploy}, "maintenance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, defaultRoleForWorkflow(appCfg, tt.flw))
		})
	}
}
