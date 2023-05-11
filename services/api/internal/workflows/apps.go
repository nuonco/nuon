package workflows

import (
	"context"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/models"

	tclient "go.temporal.io/sdk/client"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_apps.go -source=apps.go -package=workflows
func NewAppWorkflowManager(tc temporalClient) *appWorkflowManager {
	return &appWorkflowManager{
		tc: tc,
	}
}

type AppWorkflowManager interface {
	Provision(context.Context, *models.App) (string, error)
}

var _ AppWorkflowManager = (*appWorkflowManager)(nil)

type appWorkflowManager struct {
	tc temporalClient
}

func (a *appWorkflowManager) Provision(ctx context.Context, app *models.App) (string, error) {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflows.DefaultTaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":     app.OrgID,
			"app-id":     app.ID,
			"started-by": "api",
		},
	}
	args := appsv1.ProvisionRequest{
		OrgId: app.OrgID,
		AppId: app.ID,
	}

	fut, err := a.tc.ExecuteWorkflow(ctx, opts, "Provision", &args)
	if err != nil {
		return "", err
	}

	return fut.GetID(), err
}
