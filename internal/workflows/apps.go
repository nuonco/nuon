package workflows

import (
	"context"

	"github.com/powertoolsdev/api/internal/models"
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

type appProvisionWorkflowArgs struct {
	OrgID string `json:"org_id"`
	AppID string `json:"app_id"`
}

type appWorkflowManager struct {
	tc temporalClient
}

func (a *appWorkflowManager) Provision(ctx context.Context, app *models.App) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: "apps",
	}
	args := appProvisionWorkflowArgs{
		OrgID: app.OrgID.String(),
		AppID: app.ID.String(),
	}

	_, err := a.tc.ExecuteWorkflow(ctx, opts, "Provision", args)
	return err
}
