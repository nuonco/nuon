package workflows

import (
	"context"

	"github.com/powertoolsdev/api/internal/models"
	installsv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_installs.go -source=installs.go -package=workflows
const (
	sandboxVersion string = "0.10.1"
	sandboxName    string = "aws-eks"
)

func NewInstallWorkflowManager(tclient temporalClient) *installWorkflowManager {
	return &installWorkflowManager{
		tc: tclient,
	}
}

type installWorkflowManager struct {
	tc temporalClient
}

type InstallWorkflowManager interface {
	Provision(context.Context, *models.Install, string) error
	Deprovision(context.Context, *models.Install, string) error
}

var _ InstallWorkflowManager = (*installWorkflowManager)(nil)

func (i *installWorkflowManager) Provision(ctx context.Context, install *models.Install, orgID string) error {
	opts := tclient.StartWorkflowOptions{
		ID:        orgID,
		TaskQueue: "install",
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":     orgID,
			"app-id":     install.AppID.String(),
			"install-id": install.ID.String(),
		},
	}

	args := &installsv1.ProvisionRequest{
		OrgId:     orgID,
		AppId:     install.AppID.String(),
		InstallId: install.ID.String(),
		AccountSettings: &installsv1.AccountSettings{
			Region:       install.AWSSettings.Region.ToRegion(),
			AwsAccountId: "548377525120",
			AwsRoleArn:   install.AWSSettings.IamRoleArn,
		},
		SandboxSettings: &installsv1.SandboxSettings{
			Name:    sandboxName,
			Version: sandboxVersion,
		},
	}

	_, err := i.tc.ExecuteWorkflow(ctx, opts, "Provision", args)
	return err
}

func (i *installWorkflowManager) Deprovision(ctx context.Context, install *models.Install, orgID string) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: "install",
	}

	args := &installsv1.DeprovisionRequest{
		OrgId:     orgID,
		AppId:     install.AppID.String(),
		InstallId: install.ID.String(),
		AccountSettings: &installsv1.AccountSettings{
			Region:       install.AWSSettings.Region.ToRegion(),
			AwsAccountId: "548377525120",
			AwsRoleArn:   install.AWSSettings.IamRoleArn,
		},
		SandboxSettings: &installsv1.SandboxSettings{
			Name:    sandboxName,
			Version: sandboxVersion,
		},
	}

	_, err := i.tc.ExecuteWorkflow(ctx, opts, "Deprovision", args)
	return err
}
