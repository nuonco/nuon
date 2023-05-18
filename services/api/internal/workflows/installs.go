package workflows

import (
	"context"

	temporalclient "github.com/powertoolsdev/mono/pkg/clients/temporal"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_installs.go -source=installs.go -package=workflows

func NewInstallWorkflowManager(tclient temporalclient.Client) *installWorkflowManager {
	return &installWorkflowManager{
		tc: tclient,
	}
}

type installWorkflowManager struct {
	tc temporalclient.Client
}

type InstallWorkflowManager interface {
	Provision(context.Context, *models.Install, string, *models.SandboxVersion) (string, error)
	Deprovision(context.Context, *models.Install, string, *models.SandboxVersion) (string, error)
}

var _ InstallWorkflowManager = (*installWorkflowManager)(nil)

func (i *installWorkflowManager) Provision(ctx context.Context, install *models.Install, orgID string, sandboxVersion *models.SandboxVersion) (string, error) {
	opts := tclient.StartWorkflowOptions{
		ID:        orgID,
		TaskQueue: workflows.DefaultTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":     orgID,
			"app-id":     install.AppID,
			"install-id": install.ID,
			"started-by": "api",
		},
	}

	args := &installsv1.ProvisionRequest{
		OrgId:     orgID,
		AppId:     install.AppID,
		InstallId: install.ID,
		AccountSettings: &installsv1.AccountSettings{
			Region:       install.AWSSettings.Region.ToRegion(),
			AwsAccountId: "548377525120",
			AwsRoleArn:   install.AWSSettings.IamRoleArn,
		},
		SandboxSettings: &installsv1.SandboxSettings{
			Name:    sandboxVersion.SandboxName,
			Version: sandboxVersion.SandboxVersion,
		},
	}

	workflow, err := i.tc.ExecuteWorkflowInNamespace(ctx, "installs", opts, "Provision", args)
	if err != nil {
		return "", err
	}
	return workflow.GetID(), err
}

func (i *installWorkflowManager) Deprovision(ctx context.Context, install *models.Install, orgID string, sandboxVersion *models.SandboxVersion) (string, error) {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflows.DefaultTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":     orgID,
			"app-id":     install.AppID,
			"install-id": install.ID,
			"started-by": "api",
		},
	}

	args := &installsv1.DeprovisionRequest{
		OrgId:     orgID,
		AppId:     install.AppID,
		InstallId: install.ID,
		AccountSettings: &installsv1.AccountSettings{
			Region:       install.AWSSettings.Region.ToRegion(),
			AwsAccountId: "548377525120",
			AwsRoleArn:   install.AWSSettings.IamRoleArn,
		},
		SandboxSettings: &installsv1.SandboxSettings{
			Name:    sandboxVersion.SandboxName,
			Version: sandboxVersion.SandboxVersion,
		},
	}

	workflow, err := i.tc.ExecuteWorkflowInNamespace(ctx, "installs", opts, "Deprovision", args)
	if err != nil {
		return "", err
	}
	return workflow.GetID(), err
}
