package iam

import (
	"fmt"

	iamv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/iam/v1"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/roles"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type iamSet struct {
	name          string
	policyFn      func() ([]byte, error)
	iamNameFn     func() string
	trustPolicyFn func() ([]byte, error)
}

func (w *wkflow) createIAMSet(ctx workflow.Context, req *iamv1.ProvisionIAMRequest, set iamSet) (string, error) {
	l := log.With(workflow.GetLogger(ctx))
	act := NewActivities()

	l.Debug("creating %s iam policy for org %s", set.name, req.OrgId)
	policy, err := set.policyFn()
	if err != nil {
		return "", fmt.Errorf("unable to create IAM policy document: %w", err)
	}
	policyReq := CreateIAMPolicyRequest{
		AssumeRoleARN: w.cfg.OrgsIAMAccessRoleArn,
		PolicyARN: fmt.Sprintf("arn:aws:iam::%s:policy%s%s",
			w.cfg.OrgsAccountID,
			defaultIAMPath(req.OrgId),
			set.iamNameFn(),
		),
		PolicyName:     set.iamNameFn(),
		PolicyPath:     defaultIAMPath(req.OrgId),
		PolicyDocument: string(policy),
		PolicyTags:     roles.DefaultTags(req.OrgId),
	}

	policyResp, err := execCreateIAMPolicy(ctx, act, policyReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM policy: %w", err)
	}

	l.Debug("creating %s role for org %s", set.name, req.OrgId)
	trustPolicy, err := set.trustPolicyFn()
	if err != nil {
		return "", fmt.Errorf("unable to create IAM trust policy document: %w", err)
	}

	roleReq := CreateIAMRoleRequest{
		AssumeRoleARN: w.cfg.OrgsIAMAccessRoleArn,
		RoleARN: fmt.Sprintf("arn:aws:iam::%s:role%s%s",
			w.cfg.OrgsAccountID,
			defaultIAMPath(req.OrgId),
			set.iamNameFn(),
		),
		RoleName:            set.iamNameFn(),
		RolePath:            defaultIAMPath(req.OrgId),
		TrustPolicyDocument: string(trustPolicy),
		RoleTags:            roles.DefaultTags(req.OrgId),
	}
	roleResp, err := execCreateIAMRole(ctx, act, roleReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role: %w", err)
	}

	l.Debug("creating iam policy attachment for org %s", req.OrgId)
	cdpaReq := CreateIAMRolePolicyAttachmentRequest{
		AssumeRoleARN: w.cfg.OrgsIAMAccessRoleArn,
		PolicyArn:     policyResp.PolicyArn,
		RoleName:      set.iamNameFn(),
	}
	err = execCreateIAMRolePolicyAttachment(ctx, act, cdpaReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role policy attachment: %w", err)
	}

	return roleResp.RoleArn, nil
}
