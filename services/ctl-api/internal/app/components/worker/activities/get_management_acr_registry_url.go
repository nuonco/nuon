package activities

import "context"

type GetManagementACRRegistryURLRequest struct{}

// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) GetManagementACRRegistryURL(ctx context.Context, req *GetManagementACRRegistryURLRequest) (string, error) {
	return a.cfg.ManagementACRRegistryURL, nil
}
