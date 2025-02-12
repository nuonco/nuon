package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) HardDelete(ctx context.Context, orgID string) error {
	childObjs := []interface{}{
		&app.RunnerJobExecutionResult{},
		&app.RunnerJobExecution{},
		&app.RunnerJobPlan{},
		&app.RunnerJob{},
		&app.Runner{},
		&app.RunnerGroupSettings{},
		&app.RunnerGroup{},
		&app.InstallComponent{},
		&app.InstallDeploy{},
		&app.ComponentReleaseStep{},
		&app.ComponentRelease{},
		&app.ComponentBuild{},
		&app.AWSECRImageConfig{},
		&app.PublicGitVCSConfig{},
		&app.ConnectedGithubVCSConfig{},
		&app.ExternalImageComponentConfig{},
		&app.JobComponentConfig{},
		&app.DockerBuildComponentConfig{},
		&app.TerraformModuleComponentConfig{},
		&app.HelmComponentConfig{},
		&app.InstallActionWorkflowRunStep{},
		&app.InstallActionWorkflowRun{},
		&app.InstallActionWorkflow{},
		&app.ActionWorkflowTriggerConfig{},
		&app.ActionWorkflowStepConfig{},
		&app.ActionWorkflowConfig{},
		&app.ActionWorkflow{},
		&app.ComponentConfigConnection{},
		&app.ComponentDependency{},
		&app.Component{},
		&app.InstallSandboxRun{},
		&app.InstallInputs{},
		&app.InstallEvent{},
		&app.Install{},
		&app.AzureAccount{},
		&app.AWSAccount{},
		&app.AppSecret{},
		&app.AppInputConfig{},
		&app.AppInputGroup{},
		&app.AppInput{},
		&app.AppRunnerConfig{},
		&app.AppAWSDelegationConfig{},
		&app.AppSandboxConfig{},
		&app.AppConfig{},
		&app.App{},
		&app.VCSConnectionCommit{},
		&app.VCSConnection{},
		&app.InstallerMetadata{},
		&app.Installer{},
		&app.OrgHealthCheck{},
		&app.OrgInvite{},
		&app.NotificationsConfig{},
		&app.Policy{},
		&app.AccountRole{},
		&app.Role{},
	}
	for _, obj := range childObjs {
		res := h.db.WithContext(ctx).Unscoped().
			Where("org_id = ?", orgID).
			Delete(obj)
		if res.Error != nil {
			return fmt.Errorf("unable to delete %T for org: %w", obj, res.Error)
		}
	}

	// delete org
	res := h.db.WithContext(ctx).Unscoped().Delete(&app.Org{
		ID: orgID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org not found %w", gorm.ErrRecordNotFound)
	}

	return nil
}
