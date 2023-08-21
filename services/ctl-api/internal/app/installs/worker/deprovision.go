package worker

import (
	"fmt"

	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) deprovision(ctx workflow.Context, installID string, dryRun bool) error {
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		InstallID:         installID,
		Status:            "deprovisioning",
		StatusDescription: "deleting install resources - this should take about 15 minutes",
	}); err != nil {
		return fmt.Errorf("unable to update install status: %w", err)
	}

	var install app.Install
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		InstallID: installID,
	}, &install); err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	_, err := w.execDeprovisionWorkflow(ctx, dryRun, &installsv1.DeprovisionRequest{
		OrgId:     install.App.Org.ID,
		AppId:     install.App.ID,
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
		return fmt.Errorf("unable to deprovision install: %w", err)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.Delete, activities.DeleteRequest{
		InstallID: installID,
	}); err != nil {
		return fmt.Errorf("unable to delete install: %w", err)
	}
	return nil
}
