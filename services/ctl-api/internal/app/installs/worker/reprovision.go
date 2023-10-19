package worker

import (
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) reprovision(ctx workflow.Context, installID string, dryRun bool) error {
	w.updateStatus(ctx, installID, StatusProvisioning, "reprovisioning install resources")

	var install app.Install
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		InstallID: installID,
	}, &install); err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to get install from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	_, err := w.execProvisionWorkflow(ctx, dryRun, &installsv1.ProvisionRequest{
		OrgId:     install.OrgID,
		AppId:     install.AppID,
		InstallId: installID,
		AccountSettings: &installsv1.AccountSettings{
			Region:       install.AWSAccount.Region,
			AwsRoleArn:   install.AWSAccount.IAMRoleARN,
			AwsAccountId: "REMOVE - DEPRECATED",
		},
		SandboxSettings: &installsv1.SandboxSettings{
			Name:             install.SandboxRelease.Sandbox.Name,
			Version:          install.SandboxRelease.Version,
			TerraformVersion: install.SandboxRelease.TerraformVersion,
		},
	})
	if err != nil {
		accessError := credentials.ErrUnableToAssumeRole{
			RoleARN: install.AWSAccount.IAMRoleARN,
		}
		if strings.Contains(err.Error(), accessError.Error()) {
			w.updateStatus(ctx, installID, StatusAccessError, "unable to assume provided role to access account")
			return fmt.Errorf("unable to reprovision install: %w", err)
		}

		w.updateStatus(ctx, installID, StatusError, "unable to reprovision app resources")
		return fmt.Errorf("unable to provision install: %w", err)
	}

	w.updateStatus(ctx, installID, StatusActive, "app resources are provisioned")
	return nil
}
