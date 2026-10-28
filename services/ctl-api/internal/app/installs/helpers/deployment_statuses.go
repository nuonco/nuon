package helpers

import "github.com/nuonco/nuon/services/ctl-api/internal/app"

var InstallDeploymentActiveStatuses = []string{
	"",
	string(app.StatusPending),
	string(app.StatusQueued),
	string(app.StatusInProgress),
	string(app.StatusRetrying),
	string(app.AwaitingApproval),
	string(app.WorkflowStepApprovalStatusApproved),
}
