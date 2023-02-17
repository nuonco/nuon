package workflows

import (
	"context"

	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/go-common/shortid"
	appsv1 "github.com/powertoolsdev/protos/workflows/generated/types/apps/v1"
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
	Provision(context.Context, *models.App) error
}

var _ AppWorkflowManager = (*appWorkflowManager)(nil)

type appWorkflowManager struct {
	tc temporalClient
}

func (a *appWorkflowManager) Provision(ctx context.Context, app *models.App) error {
	orgID := shortid.ParseUUID(app.OrgID)
	appID := shortid.ParseUUID(app.ID)

	opts := tclient.StartWorkflowOptions{
		TaskQueue: "apps",
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":     orgID,
			"app-id":     appID,
			"started-by": "api",
		},
	}
	args := appsv1.ProvisionRequest{
		OrgId: orgID,
		AppId: appID,
	}

	_, err := a.tc.ExecuteWorkflow(ctx, opts, "Provision", &args)
	return err
}
