package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
