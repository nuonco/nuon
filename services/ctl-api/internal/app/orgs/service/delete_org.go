package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID DeleteOrg
// @Summary	Delete an org
// @Schemes
// @Description.markdown	delete_org.md
// @Tags			orgs
// @Accept			json
// @Security APIKey
// @Security OrgID
// @Produce		json
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{boolean}	ok
// @Router			/v1/orgs/current [DELETE]
func (s *service) DeleteOrg(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if org.OrgType == app.OrgTypeIntegration {
		err := s.deleteIntegrationOrg(ctx, org.ID)
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.JSON(http.StatusOK, true)
		return
	}

	err = s.deleteOrg(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type: sigs.OperationDelete,
	})
	ctx.JSON(http.StatusOK, true)
}

func (s *service) deleteIntegrationOrg(ctx context.Context, orgID string) error {
	deleteObjs := []interface{}{
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
	for _, obj := range deleteObjs {
		res := s.db.WithContext(ctx).Unscoped().
			Where("org_id = ?", orgID).
			Delete(obj)
		if res.Error != nil {
			return fmt.Errorf("unable to delete %T for org: %w", obj, res.Error)
		}
	}

	// delete org
	res := s.db.WithContext(ctx).Unscoped().Delete(&app.Org{
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

func (s *service) deleteOrg(ctx context.Context, orgID string) error {
	org := app.Org{
		ID: orgID,
	}
	res := s.db.WithContext(ctx).Model(&org).Updates(app.Org{
		StatusDescription: "delete has been queued",
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org not found: %w", gorm.ErrRecordNotFound)
	}
	return nil
}
