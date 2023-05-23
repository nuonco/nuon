package activities

import "context"

type SyncArtifactsResponse struct {
	WorkflowID string
}

func (a *activities) SyncArtifactsJob(ctx context.Context, buildID string, installID string) (*SyncArtifactsResponse, error) {
	return &SyncArtifactsResponse{}, nil
}
