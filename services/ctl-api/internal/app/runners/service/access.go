package service

import (
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/access"
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
