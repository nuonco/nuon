package iam

import (
	"fmt"
	"time"

	iamv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1/iam/v1"
	workers "github.com/powertoolsdev/workers-orgs/internal"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultActivityTimeout time.Duration = time.Second * 10
	defaultIAMPolicyPath   string        = "/orgs/"
	defaultIAMRolePath     string        = "/orgs/"
)

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg workers.Config) wkflow {
	return wkflow{
		cfg: cfg,
	}
}

type wkflow struct {
	cfg workers.Config
}

// ProvisionIAM is a workflow that creates org specific IAM roles in the designated orgs IAM account
func (w wkflow) ProvisionIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (*iamv1.ProvisionIAMResponse, error) {
	resp := &iamv1.ProvisionIAMResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	// get waypoint server cookie
	l := log.With(workflow.GetLogger(ctx))
	act := NewActivities()

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	// create deployments
	l.Debug("creating deployments iam policy for org %s", req.OrgId)
	deploymentsPolicy, err := deploymentsIAMPolicy(w.cfg.OrgDeploymentsBucketName, req.OrgId)
	if err != nil {
		return resp, fmt.Errorf("unable to create deployments IAM policy document: %w", err)
	}
	cdpReq := CreateIAMPolicyRequest{
		OrgsIAMAccessRoleArn: w.cfg.OrgsIAMAccessRoleArn,
		PolicyName:           deploymentsIAMName(req.OrgId),
		PolicyPath:           defaultIAMPolicyPath,
		PolicyDocument:       string(deploymentsPolicy),
	}
	cdpResp, err := execCreateIAMPolicy(ctx, act, cdpReq)
	if err != nil {
		return resp, fmt.Errorf("unable to create deployments IAM policy: %w", err)
	}

	l.Debug("creating deployments iam rol for org %s", req.OrgId)
	deploymentsTrustPolicy, err := deploymentsIAMTrustPolicy(w.cfg)
	if err != nil {
		return resp, fmt.Errorf("unable to create deployments IAM trust policy document: %w", err)
	}
	cdrReq := CreateIAMRoleRequest{
		OrgsIAMAccessRoleArn: w.cfg.OrgsIAMAccessRoleArn,
		RoleName:             deploymentsIAMName(req.OrgId),
		RolePath:             defaultIAMRolePath,
		TrustPolicyDocument:  string(deploymentsTrustPolicy),
		RoleTags:             defaultTags(req.OrgId),
	}
	cdrResp, err := execCreateIAMRole(ctx, act, cdrReq)
	if err != nil {
		return resp, fmt.Errorf("unable to create deployments IAM role: %w", err)
	}

	l.Debug("creating deployments iam policy attachment for org %s", req.OrgId)
	cdpaReq := CreateIAMRolePolicyAttachmentRequest{
		OrgsIAMAccessRoleArn: w.cfg.OrgsIAMAccessRoleArn,
		PolicyArn:            cdpResp.PolicyArn,
		RoleArn:              cdrResp.RoleArn,
	}
	_, err = execCreateIAMRolePolicyAttachment(ctx, act, cdpaReq)
	if err != nil {
		return resp, fmt.Errorf("unable to create deployments IAM role policy attachment: %w", err)
	}

	return resp, nil
}

func execCreateIAMPolicy(
	ctx workflow.Context,
	act *Activities,
	req CreateIAMPolicyRequest,
) (CreateIAMPolicyResponse, error) {
	var resp CreateIAMPolicyResponse

	l := workflow.GetLogger(ctx)

	l.Debug("executing create iam policy activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateIAMPolicy, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execCreateIAMRole(
	ctx workflow.Context,
	act *Activities,
	req CreateIAMRoleRequest,
) (CreateIAMRoleResponse, error) {
	var resp CreateIAMRoleResponse

	l := workflow.GetLogger(ctx)

	l.Debug("executing create iam role activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateIAMRole, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execCreateIAMRolePolicyAttachment(
	ctx workflow.Context,
	act *Activities,
	req CreateIAMRolePolicyAttachmentRequest,
) (CreateIAMRolePolicyAttachmentResponse, error) {
	var resp CreateIAMRolePolicyAttachmentResponse

	l := workflow.GetLogger(ctx)

	l.Debug("executing create iam role policy attachment activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateIAMRolePolicyAttachment, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
