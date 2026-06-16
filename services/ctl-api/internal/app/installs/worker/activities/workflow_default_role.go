package activities

import "context"

type WorkflowDefaultRoleEnabledRequest struct{}

// @temporal-gen-v2 activity
func (a *Activities) WorkflowDefaultRoleEnabled(ctx context.Context, req WorkflowDefaultRoleEnabledRequest) (bool, error) {
	return a.cfg.WorkflowDefaultRoleEnabled, nil
}
