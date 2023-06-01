package activities

import (
	"context"

	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	meta "github.com/powertoolsdev/mono/pkg/workflows/meta"
)

// StartStartRequest logs to S3 that a workflow run was started, along with the request object that was given to the workflow.
func (a *Activities) StartStartRequest(ctx context.Context, req *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
	act := meta.NewStartActivity()
	return act.StartRequest(ctx, req)
}

// FinishStartRequest logs to S3 that a workflow run was finished, along with the response object returned by the workflow.
func (a *Activities) FinishStartRequest(ctx context.Context, req *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
	act := meta.NewFinishActivity()
	return act.FinishRequest(ctx, req)
}
