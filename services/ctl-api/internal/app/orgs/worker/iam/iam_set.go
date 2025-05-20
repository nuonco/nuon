package orgiam

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/iam/roles"
)

type iamSet struct {
	name          string
	policyFn      func() ([]byte, error)
	iamNameFn     func() string
	trustPolicyFn func() ([]byte, error)
}

func (w *Wkflow) createIAMSet(ctx workflow.Context, req *ProvisionIAMRequest, set iamSet) (string, error) {
	l := log.With(workflow.GetLogger(ctx))

	l.Debug("creating %s iam policy for org %s", set.name, req.OrgID)
	policy, err := set.policyFn()
	if err != nil {
		return "", fmt.Errorf("unable to create IAM policy document: %w", err)
	}
	policyReq := CreateIAMPolicyRequest{
		AssumeRoleARN: w.cfg.ManagementIAMRoleARN,
		PolicyARN: fmt.Sprintf("arn:aws:iam::%s:policy%s%s",
			w.cfg.ManagementAccountID,
			defaultIAMPath(req.OrgID),
			set.iamNameFn(),
		),
		PolicyName:     set.iamNameFn(),
		PolicyPath:     defaultIAMPath(req.OrgID),
		PolicyDocument: string(policy),
		PolicyTags:     roles.DefaultTags(req.OrgID),
	}

	policyResp, err := AwaitCreateIAMPolicy(ctx, policyReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM policy: %w", err)
	}

	l.Debug("creating %s role for org %s", set.name, req.OrgID)
	trustPolicy, err := set.trustPolicyFn()
	if err != nil {
		return "", fmt.Errorf("unable to create IAM trust policy document: %w", err)
	}

	roleReq := CreateIAMRoleRequest{
		AssumeRoleARN: w.cfg.ManagementIAMRoleARN,
		RoleARN: fmt.Sprintf("arn:aws:iam::%s:role%s%s",
			w.cfg.ManagementAccountID,
			defaultIAMPath(req.OrgID),
			set.iamNameFn(),
		),
		RoleName:            set.iamNameFn(),
		RolePath:            defaultIAMPath(req.OrgID),
		TrustPolicyDocument: string(trustPolicy),
		RoleTags:            roles.DefaultTags(req.OrgID),
	}
	roleResp, err := AwaitCreateIAMRole(ctx, roleReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role: %w", err)
	}

	l.Debug("creating iam policy attachment for org %s", req.OrgID)
	cdpaReq := CreateIAMRolePolicyAttachmentRequest{
		AssumeRoleARN: w.cfg.ManagementIAMRoleARN,
		PolicyArn:     policyResp.PolicyArn,
		RoleName:      set.iamNameFn(),
	}
	_, err = AwaitCreateIAMRolePolicyAttachment(ctx, cdpaReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role policy attachment: %w", err)
	}

	return roleResp.RoleArn, nil
}
