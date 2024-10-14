package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

func (s *Helpers) ContextFromJob(ctx context.Context, jobID string) (context.Context, error) {
	job, err := s.getJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("unable to get job: %w", err)
	}

	ctx = cctx.SetOrgIDContext(ctx, job.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, job.CreatedByID)

	return ctx, nil
}
