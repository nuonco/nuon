package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) jobResources(ctx *gin.Context) (*authz.Resources, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return access.Jobs(s.db, orgID)
}

func (s *service) jobTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	job, err := s.getOrgRunnerJob(ctx, ctx.Param("runner_job_id"), orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return access.JobTarget(ctx, s.db, *job)
}

func (s *service) logStreamTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	stream, err := s.getOrgLogStream(ctx, ctx.Param("log_stream_id"), orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return access.LogStreamTarget(ctx, s.db, *stream)
}

func (s *service) runnerTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := access.Runners.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx.Request.Context(), ctx.Param("runner_id"))
}

func (s *service) processRunnerTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var process app.RunnerProcess
	if err := s.db.WithContext(ctx.Request.Context()).Where(app.RunnerProcess{ID: ctx.Param("process_id"), OrgID: orgID}).Take(&process).Error; err != nil {
		return permissions.Target{}, err
	}
	if runnerID := ctx.Param("runner_id"); runnerID != "" && runnerID != process.RunnerID {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	if shutdownID := ctx.Param("shutdown_id"); shutdownID != "" {
		var shutdown app.RunnerProcessShutdown
		if err := s.db.WithContext(ctx.Request.Context()).Where(app.RunnerProcessShutdown{
			ID: shutdownID, OrgID: orgID, RunnerProcessID: process.ID,
		}).Take(&shutdown).Error; err != nil {
			return permissions.Target{}, err
		}
	}
	resources, err := access.Runners.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx.Request.Context(), process.RunnerID)
}

func (s *service) terraformResources(ctx *gin.Context) (*authz.Resources, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return access.TerraformWorkspaces(s.db, orgID)
}

func (s *service) terraformTarget(ctx *gin.Context) (permissions.Target, error) {
	resources, err := s.terraformResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	id := ctx.Param("workspace_id")
	if id == "" {
		id = ctx.Query("workspace_id")
	}
	return resources.Resolve(ctx.Request.Context(), id)
}

func (s *service) terraformCreateTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var req CreateTerraformWorkspaceRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return permissions.Target{}, stderr.NewInvalidRequest(err)
	}
	return access.TerraformOwnerTarget(ctx.Request.Context(), s.db, orgID, req.OwnerType, req.OwnerID)
}

func (s *service) terraformStateTarget(ctx *gin.Context) (permissions.Target, error) {
	target, err := s.terraformTarget(ctx)
	if err != nil {
		return target, err
	}
	if id := ctx.Query("job_id"); id != "" {
		var job app.RunnerJob
		if err := s.db.WithContext(ctx.Request.Context()).Where(app.RunnerJob{ID: id, OrgID: target.OrgID}).Take(&job).Error; err != nil {
			return target, err
		}
		jobTarget, err := access.JobTarget(ctx.Request.Context(), s.db, job)
		if err != nil {
			return target, err
		}
		if jobTarget.AppID != target.AppID || jobTarget.InstallID != target.InstallID {
			return target, gorm.ErrRecordNotFound
		}
	}
	return target, nil
}

func (s *service) protocolJobTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	id := ctx.Param("runner_job_id")
	if id == "" {
		id = ctx.Param("job_id")
	}
	if id == "" {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	var job app.RunnerJob
	if err := s.db.WithContext(ctx.Request.Context()).Where(app.RunnerJob{ID: id, OrgID: orgID}).Take(&job).Error; err != nil {
		return permissions.Target{}, err
	}
	if runnerID := ctx.Param("runner_id"); runnerID != "" && runnerID != job.RunnerID {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	if executionID := ctx.Param("runner_job_execution_id"); executionID != "" {
		var execution app.RunnerJobExecution
		if err := s.db.WithContext(ctx.Request.Context()).Where(app.RunnerJobExecution{
			ID: executionID, OrgID: orgID, RunnerJobID: job.ID,
		}).Take(&execution).Error; err != nil {
			return permissions.Target{}, err
		}
	}
	target, err := access.JobTarget(ctx.Request.Context(), s.db, job)
	if err != nil {
		return permissions.Target{}, err
	}
	runners, err := access.Runners.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	runner, err := runners.Resolve(ctx.Request.Context(), job.RunnerID)
	if err != nil {
		return permissions.Target{}, err
	}
	if runner.InstallID != "" && (runner.InstallID != target.InstallID || runner.AppID != target.AppID) {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	return target, nil
}

func (s *service) heartbeatRunnerTarget(ctx *gin.Context) (permissions.Target, error) {
	target, err := s.runnerTarget(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var req CreateRunnerHeartBeatRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return permissions.Target{}, stderr.NewInvalidRequest(err)
	}
	if req.ProcessID != "" {
		var process app.RunnerProcess
		if err := s.db.WithContext(ctx.Request.Context()).Where(app.RunnerProcess{
			ID: req.ProcessID, OrgID: target.OrgID, RunnerID: target.ID,
		}).Take(&process).Error; err != nil {
			return permissions.Target{}, err
		}
	}
	return target, nil
}

func (s *service) componentHealthRunnerTarget(ctx *gin.Context) (permissions.Target, error) {
	target, err := s.runnerTarget(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var req CreateComponentHealthRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return permissions.Target{}, stderr.NewInvalidRequest(err)
	}
	if target.InstallID == "" || target.InstallID != req.InstallID {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	for _, component := range req.Components {
		if component.InstallComponentID == "" {
			return permissions.Target{}, gorm.ErrRecordNotFound
		}
		var row app.InstallComponent
		if err := s.db.WithContext(ctx.Request.Context()).Where(app.InstallComponent{
			ID: component.InstallComponentID, OrgID: target.OrgID, InstallID: target.InstallID,
			ComponentID: component.ComponentID,
		}).Take(&row).Error; err != nil {
			return permissions.Target{}, err
		}
		if _, err := access.TerraformOwnerTarget(ctx.Request.Context(), s.db, target.OrgID, "install_components", row.ID); err != nil {
			return permissions.Target{}, err
		}
	}
	return target, nil
}

func (s *service) telemetryRunnerTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	runners, err := access.Runners.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return runners.Resolve(ctx.Request.Context(), acct.Subject)
}

func (s *service) helmTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var chart app.HelmChart
	if err := s.db.WithContext(ctx.Request.Context()).Where(app.HelmChart{
		ID: ctx.Param("helm_chart_id"), OrgID: orgID,
	}).Take(&chart).Error; err != nil {
		return permissions.Target{}, err
	}
	target, err := access.TerraformOwnerTarget(ctx.Request.Context(), s.db, orgID, chart.OwnerType, chart.OwnerID)
	if err != nil {
		return permissions.Target{}, err
	}
	target.ID = chart.ID
	return target, nil
}

func (s *service) installComponentTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var component app.InstallComponent
	if err := s.db.WithContext(ctx.Request.Context()).Where(app.InstallComponent{
		InstallID: ctx.Param("install_id"), ComponentID: ctx.Param("component_id"), OrgID: orgID,
	}).Take(&component).Error; err != nil {
		return permissions.Target{}, err
	}
	if _, err := access.TerraformOwnerTarget(ctx.Request.Context(), s.db, orgID, "install_components", component.ID); err != nil {
		return permissions.Target{}, err
	}
	installs, err := installaccess.Installs.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return installs.Resolve(ctx.Request.Context(), component.InstallID)
}
