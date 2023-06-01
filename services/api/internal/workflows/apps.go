package workflows

import (
	"context"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/models"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_apps.go -source=apps.go -package=workflows
func NewAppWorkflowManager(workflowsClient workflows.Client) *appWorkflowManager {
	return &appWorkflowManager{
		workflowsClient: workflowsClient,
	}
}

type AppWorkflowManager interface {
	Provision(context.Context, *models.App) (string, error)
}

var _ AppWorkflowManager = (*appWorkflowManager)(nil)

type appWorkflowManager struct {
	workflowsClient workflows.Client
}

func (a *appWorkflowManager) Provision(ctx context.Context, app *models.App) (string, error) {
	args := appsv1.ProvisionRequest{
		OrgId: app.OrgID,
		AppId: app.ID,
	}
	return a.workflowsClient.TriggerAppProvision(ctx, &args)
}
