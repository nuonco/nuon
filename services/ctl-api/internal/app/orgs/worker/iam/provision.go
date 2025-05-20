package orgiam

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/iam/roles"
)

func (w Wkflow) provisionRunnerIAM(ctx workflow.Context, req *ProvisionIAMRequest) (string, error) {
	set := iamSet{
		name: "runner",
		policyFn: func() ([]byte, error) {
			return roles.RunnerIAMPolicy(w.cfg.ManagementECRRegistryARN, req.OrgID)
		},
		iamNameFn: func() string {
			return roles.RunnerIAMName(req.OrgID)
		},
		trustPolicyFn: func() ([]byte, error) {
			return roles.RunnerIAMTrustPolicy(w.cfg.OrgRunnerSupportRoleARN, w.cfg.OrgRunnerOIDCProviderARN,
				w.cfg.OrgRunnerOIDCProviderURL, req.OrgID)
		},
	}
	return w.createIAMSet(ctx, req, set)
}

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-callback ProvisionIAMCallback
func (w Wkflow) ProvisionIAM(ctx workflow.Context, req *ProvisionIAMRequest) (*ProvisionIAMResponse, error) {
	resp := &ProvisionIAMResponse{}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	runnerRoleArn, err := w.provisionRunnerIAM(ctx, req)
	if err != nil {
		return resp, fmt.Errorf("unable to provision runner IAM role: %w", err)
	}
	resp.RunnerRoleArn = runnerRoleArn

	return resp, nil
}
