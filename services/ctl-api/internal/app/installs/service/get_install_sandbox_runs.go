package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID GetInstallSandboxRuns
// @Summary	get an installs sandbox runs
// @Description.markdown	 get_install_sandbox_runs.md
// @Param			install_id	path	string	true	"install ID"
// @Param   offset query int	 false	"offset of results to return"	Default(0)
// @Param   limit  query int	 false	"limit of results to return"	     Default(10)
// @Param   x-nuon-pagination-enabled header bool false "Enable pagination"
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
// @Success		200				{array}		app.InstallSandboxRun
// @Router			/v1/installs/{install_id}/sandbox-runs [GET]
func (s *service) GetInstallSandboxRuns(ctx *gin.Context) {
	appID := ctx.Param("install_id")

	installSandboxRuns, err := s.getInstallSandboxRuns(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install sandbox runs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installSandboxRuns)
}

func (s *service) getInstallSandboxRuns(ctx *gin.Context, installID string) ([]app.InstallSandboxRun, error) {
	var installSandboxRuns []app.InstallSandboxRun
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithPagination).
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("ActionWorkflowRuns").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("RunnerJob").
		Preload("LogStream").
		Preload("CreatedBy").
		Where("install_id = ?", installID).
		Order("created_at desc").
		Find(&installSandboxRuns)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install sandbox runs: %w", res.Error)
	}

	installSandboxRuns, err := db.HandlePaginatedResponse(ctx, installSandboxRuns)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return installSandboxRuns, nil
}
