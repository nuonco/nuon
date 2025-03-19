package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetInstallSandboxRun
// @Summary	get an install sandbox run
// @Description.markdown	 get_install_sandbox_run.md
// @Param			run_id	path	string	true	"run ID"
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}		app.InstallSandboxRun
// @Router			/v1/installs/sandbox-runs/{run_id} [GET]
func (s *service) GetInstallSandboxRun(ctx *gin.Context) {
	runID := ctx.Param("run_id")

	installSandboxRun, err := s.getInstallSandboxRun(ctx, runID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install sandbox run: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installSandboxRun)
}

func (s *service) getInstallSandboxRun(ctx *gin.Context, runID string) (*app.InstallSandboxRun, error) {
	var installSandboxRun app.InstallSandboxRun
	res := s.db.WithContext(ctx).
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("ActionWorkflowRuns").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("RunnerJob").
		Preload("LogStream").
		Where(app.InstallSandboxRun{
			ID: runID,
		}).
		First(&installSandboxRun)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install sandbox run: %w", res.Error)
	}

	return &installSandboxRun, nil
}
