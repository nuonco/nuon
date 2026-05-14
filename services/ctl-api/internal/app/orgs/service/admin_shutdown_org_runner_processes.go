package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type GracefulShutdownOrgRunnerProcessesRequest struct{}

// @ID						AdminGracefulShutdownOrgRunnerProcesses
// @Summary				graceful shutdown all active runner processes for an org
// @Description.markdown	admin_graceful_shutdown_org_runner_processes.md
// @Param					org_id	path	string										true	"org ID"
// @Param					req		body	GracefulShutdownOrgRunnerProcessesRequest	true	"Input"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/orgs/{org_id}/admin-graceful-shutdown-processes [POST]
func (s *service) AdminGracefulShutdownOrgRunnerProcesses(ctx *gin.Context) {
	s.shutdownOrgRunnerProcesses(ctx, app.RunnerProcessShutdownTypeGraceful)
}

type ForceShutdownOrgRunnerProcessesRequest struct{}

// @ID						AdminForceShutdownOrgRunnerProcesses
// @Summary				force shutdown all active runner processes for an org
// @Description.markdown	admin_force_shutdown_org_runner_processes.md
// @Param					org_id	path	string									true	"org ID"
// @Param					req		body	ForceShutdownOrgRunnerProcessesRequest	true	"Input"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/orgs/{org_id}/admin-force-shutdown-processes [POST]
func (s *service) AdminForceShutdownOrgRunnerProcesses(ctx *gin.Context) {
	s.shutdownOrgRunnerProcesses(ctx, app.RunnerProcessShutdownTypeForce)
}

func (s *service) shutdownOrgRunnerProcesses(ctx *gin.Context, shutdownType app.RunnerProcessShutdownType) {
	orgID := ctx.Param("org_id")

	// Accept an empty body
	if err := ctx.ShouldBindJSON(&struct{}{}); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	var processes []app.RunnerProcess
	if res := s.db.WithContext(ctx).
		Where(app.RunnerProcess{OrgID: org.ID}).
		Where("composite_status::jsonb ->> 'status' IN ('active', 'offline')").
		Find(&processes); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get org runner processes: %w", res.Error))
		return
	}

	for i := range processes {
		if _, err := s.runnersHelpers.ShutdownProcess(ctx, &processes[i], shutdownType); err != nil {
			s.l.Warn("unable to shutdown runner process",
				zap.String("process_id", processes[i].ID),
				zap.Error(err),
			)
		}
	}

	ctx.JSON(http.StatusOK, true)
}
