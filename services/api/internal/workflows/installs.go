package workflows

import (
	"context"

	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/models"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_installs.go -source=installs.go -package=workflows

func NewInstallWorkflowManager(workflowsClient workflows.Client) *installWorkflowManager {
	return &installWorkflowManager{workflowsClient}
}

type installWorkflowManager struct {
	workflowsClient workflows.Client
}

type InstallWorkflowManager interface {
	Provision(context.Context, *models.Install, string, *models.SandboxVersion) (string, error)
	Deprovision(context.Context, *models.Install, string, *models.SandboxVersion) (string, error)
}

var _ InstallWorkflowManager = (*installWorkflowManager)(nil)

func (i *installWorkflowManager) Provision(ctx context.Context, install *models.Install, orgID string, sandboxVersion *models.SandboxVersion) (string, error) {
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

	workflowID, err := i.workflowsClient.TriggerInstallProvision(ctx, args)
	if err != nil {
		return "", err
	}
	return workflowID, err
}

func (i *installWorkflowManager) Deprovision(ctx context.Context, install *models.Install, orgID string, sandboxVersion *models.SandboxVersion) (string, error) {
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

	workflowID, err := i.workflowsClient.TriggerInstallDeprovision(ctx, args)
	if err != nil {
		return "", err
	}
	return workflowID, err
}
