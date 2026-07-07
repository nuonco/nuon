package activities

import "context"

type GetManagementIAMRoleARNRequest struct{}

// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) GetManagementIAMRoleARN(ctx context.Context, req *GetManagementIAMRoleARNRequest) (string, error) {
	return a.cfg.ManagementIAMRoleARN, nil
}
