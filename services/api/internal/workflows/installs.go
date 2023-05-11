package workflows

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_installs.go -source=installs.go -package=workflows

func NewInstallWorkflowManager(tclient temporalClient) *installWorkflowManager {
	return &installWorkflowManager{
		tc: tclient,
	}
}

type installWorkflowManager struct {
	tc temporalClient
}

type InstallWorkflowManager interface {
	Provision(context.Context, *models.Install, string, *models.SandboxVersion) (string, error)
	Deprovision(context.Context, *models.Install, string, *models.SandboxVersion) (string, error)
}

var _ InstallWorkflowManager = (*installWorkflowManager)(nil)

func (i *installWorkflowManager) Provision(ctx context.Context, install *models.Install, orgID string, sandboxVersion *models.SandboxVersion) (string, error) {
	installID, err := shortid.ParseString(install.ID.String())
	if err != nil {
		return "", fmt.Errorf("unable to parse ids to shortids: %w", err)
	}

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
			"install-id": installID,
			"started-by": "api",
		},
	}

	args := &installsv1.ProvisionRequest{
		OrgId:     orgID,
		AppId:     install.AppID,
		InstallId: installID,
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

	workflow, err := i.tc.ExecuteWorkflow(ctx, opts, "Provision", args)
	if err != nil {
		return "", err
	}
	return workflow.GetID(), err
}

func (i *installWorkflowManager) Deprovision(ctx context.Context, install *models.Install, orgID string, sandboxVersion *models.SandboxVersion) (string, error) {
	installID, err := shortid.ParseString(install.ID.String())
	if err != nil {
		return "", fmt.Errorf("unable to parse ids to shortids: %w", err)
	}

	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflows.DefaultTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":     orgID,
			"app-id":     install.AppID,
			"install-id": installID,
			"started-by": "api",
		},
	}

	args := &installsv1.DeprovisionRequest{
		OrgId:     orgID,
		AppId:     install.AppID,
		InstallId: installID,
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

	workflow, err := i.tc.ExecuteWorkflow(ctx, opts, "Deprovision", args)
	if err != nil {
		return "", err
	}
	return workflow.GetID(), err
}
