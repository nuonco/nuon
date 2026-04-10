package orgiam

import (
	"fmt"
	"strings"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/iam/roles"
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

// @temporal-gen-v2 workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-generator ProvisionIAMCallback
func (w Wkflow) ProvisionIAM(ctx workflow.Context, req *ProvisionIAMRequest) (*ProvisionIAMResponse, error) {
	resp := &ProvisionIAMResponse{}

	// GCP uses Workload Identity — create GCP service account + binding.
	if w.cfg.IsGCP() {
		activityOpts := workflow.ActivityOptions{
			ScheduleToCloseTimeout: defaultActivityTimeout,
		}
		ctx = workflow.WithActivityOptions(ctx, activityOpts)

		gcpReq := CreateGCPServiceAccountRequest{
			ProjectID:             w.cfg.ManagementAccountID,
			OrgID:                 req.OrgID,
			K8sNamespace:          req.RunnerID,
			K8sServiceAccountName: fmt.Sprintf("runner-%s", req.OrgID),
		}
		_, err := AwaitCreateGCPServiceAccount(ctx, &gcpReq)
		if err != nil {
			return resp, fmt.Errorf("unable to provision GCP service account: %w", err)
		}
		return resp, nil
	}

	// Azure uses AKS Workload Identity — create per-org managed identity + binding.
	if w.cfg.IsAzure() {
		activityOpts := workflow.ActivityOptions{
			ScheduleToCloseTimeout: defaultActivityTimeout,
		}
		ctx = workflow.WithActivityOptions(ctx, activityOpts)

		azureReq := CreateAzureManagedIdentityRequest{
			SubscriptionID:        w.cfg.ManagementAzureSubscriptionID,
			ResourceGroup:         w.cfg.ManagementAzureResourceGroup,
			TenantID:              w.cfg.ManagementAzureTenantID,
			OrgID:                 req.OrgID,
			Location:              w.cfg.OrgRunnerRegion,
			AKSOIDCIssuerURL:      w.cfg.ManagementAzureOIDCIssuerURL,
			K8sNamespace:          req.RunnerID,
			K8sServiceAccountName: fmt.Sprintf("runner-%s", req.OrgID),
			ACRResourceID: fmt.Sprintf(
				"/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ContainerRegistry/registries/%s",
				w.cfg.ManagementAzureSubscriptionID,
				w.cfg.ManagementAzureResourceGroup,
				strings.TrimSuffix(strings.Split(w.cfg.ManagementACRRegistryURL, ".")[0], ".azurecr"),
			),
		}
		azureResp, err := AwaitCreateAzureManagedIdentity(ctx, &azureReq)
		if err != nil {
			return resp, fmt.Errorf("unable to provision Azure managed identity: %w", err)
		}
		resp.AzureClientID = azureResp.ClientID
		return resp, nil
	}

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
