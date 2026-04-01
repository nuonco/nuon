package activities

import "context"

type GetCloudProviderRequest struct{}

// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) GetCloudProvider(ctx context.Context, req *GetCloudProviderRequest) (string, error) {
	return a.cfg.CloudProvider, nil
}
