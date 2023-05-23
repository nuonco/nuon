package activities

import "context"

type DeployInstanceResponse struct {
	WorkflowID string
}

func (a *activities) DeployInstanceJob(ctx context.Context, instanceID string) (*DeployInstanceResponse, error) {
	return &DeployInstanceResponse{}, nil
}
