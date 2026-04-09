package orgiam

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/iam/roles"
)

// @temporal-gen-v2 workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-generator DeprovisionIAMCallback
func (w Wkflow) DeprovisionIAM(ctx workflow.Context, req *DeprovisionIAMRequest) (*DeprovisionIAMResponse, error) {
	resp := &DeprovisionIAMResponse{}

	// GCP uses Workload Identity — no AWS IAM roles to deprovision.
	if w.cfg.IsGCP() {
		return resp, nil
	}

	// Azure — delete the per-org managed identity (cascades federated creds + role assignments).
	if w.cfg.IsAzure() {
		_, err := AwaitDeleteAzureManagedIdentity(ctx, &DeleteAzureManagedIdentityRequest{
			SubscriptionID: w.cfg.ManagementAzureSubscriptionID,
			ResourceGroup:  w.cfg.ManagementAzureResourceGroup,
			OrgID:          req.OrgID,
		})
		if err != nil {
			return resp, fmt.Errorf("unable to delete Azure managed identity: %w", err)
		}
		return resp, nil
	}

	status := make(map[string]interface{})
	nameFns := map[string]func(string) string{
		"runners": roles.RunnerIAMName,
	}
	for step, nameFn := range nameFns {
		if err := w.execDeprovisionRole(ctx,
			req,
			nameFn); err != nil {

			status[step] = fmt.Errorf("unable to delete IAM role: %w", err).Error()
			continue
		}
		status[step] = "ok"
	}

	_, err := structpb.NewStruct(status)
	if err != nil {
		return resp, fmt.Errorf("unable to convert struct to proto: %w", err)
	}

	return resp, nil
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Wkflow) execDeprovisionRole(ctx workflow.Context,
	req *DeprovisionIAMRequest,
	nameFn func(string) string,
) error {
	policyARN := fmt.Sprintf("arn:aws:iam::%s:policy%s%s", w.cfg.ManagementAccountID, defaultIAMPath(req.OrgID), nameFn(req.OrgID))

	deleteAttachmentReq := DeleteIAMRolePolicyAttachmentRequest{
		AssumeRoleARN: w.cfg.ManagementIAMRoleARN,
		PolicyArn:     policyARN,
		RoleName:      nameFn(req.OrgID),
	}
	_, attachmentErr := AwaitDeleteIAMRolePolicyAttachment(ctx, deleteAttachmentReq)

	deleteRoleReq := DeleteIAMRoleRequest{
		AssumeRoleARN: w.cfg.ManagementIAMRoleARN,
		RoleName:      nameFn(req.OrgID),
	}
	_, roleErr := AwaitDeleteIAMRole(ctx, deleteRoleReq)

	deletePolicyReq := DeleteIAMPolicyRequest{
		AssumeRoleARN: w.cfg.ManagementIAMRoleARN,
		PolicyARN:     policyARN,
	}
	_, policyErr := AwaitDeleteIAMPolicy(ctx, deletePolicyReq)

	return firstError(attachmentErr, roleErr, policyErr)
}
