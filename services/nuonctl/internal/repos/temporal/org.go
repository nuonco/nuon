package temporal

import (
	"context"
	"fmt"

	orgsv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/orgs/v1"
	tclient "go.temporal.io/sdk/client"
)

// TODO(jm): eventually rename this workflow to Provision
func (r *repo) TriggerOrgSignup(ctx context.Context, req *orgsv1.SignupRequest) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: "org",
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"started-by": "nuonctl",
		},
	}

	_, err := r.Client.ExecuteWorkflow(ctx, opts, "Signup", req)
	if err != nil {
		return fmt.Errorf("unable to start deployment: %w", err)
	}

	return nil
}

func (r *repo) ExecOrgSignup(ctx context.Context, req *orgsv1.SignupRequest) (*orgsv1.SignupResponse, error) {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: "org",
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"started-by": "nuonctl",
		},
	}

	resp := &orgsv1.SignupResponse{}
	fut, err := r.Client.ExecuteWorkflow(ctx, opts, "Signup", req)
	if err != nil {
		return nil, fmt.Errorf("unable to start signup: %w", err)
	}

	if err := fut.Get(ctx, resp); err != nil {
		return nil, fmt.Errorf("unable to get response: %w", err)
	}

	return resp, nil
}
