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

func (w wkflow) provisionInstallationsIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) error {
	l := log.With(workflow.GetLogger(ctx))
	act := NewActivities()

	// create deployments
	l.Debug("creating deployments iam policy for org %s", req.OrgId)
	installationsPolicy, err := installationsIAMPolicy(w.cfg.OrgInstallationsBucketName, req.OrgId)
	if err != nil {
		return fmt.Errorf("unable to create IAM policy document: %w", err)
	}
	cdpReq := CreateIAMPolicyRequest{
		OrgsIAMAccessRoleArn: w.cfg.OrgsIAMAccessRoleArn,
		PolicyName:           installationsIAMName(req.OrgId),
		PolicyPath:           defaultIAMPolicyPath,
		PolicyDocument:       string(installationsPolicy),
	}
	cdpResp, err := execCreateIAMPolicy(ctx, act, cdpReq)
	if err != nil {
		return fmt.Errorf("unable to create IAM policy: %w", err)
	}

	l.Debug("creating iam role for org %s", req.OrgId)
	installationsTrustPolicy, err := installationsIAMTrustPolicy(w.cfg)
	if err != nil {
		return fmt.Errorf("unable to create IAM trust policy document: %w", err)
	}
	cdrReq := CreateIAMRoleRequest{
		OrgsIAMAccessRoleArn: w.cfg.OrgsIAMAccessRoleArn,
		RoleName:             installationsIAMName(req.OrgId),
		RolePath:             defaultIAMRolePath,
		TrustPolicyDocument:  string(installationsTrustPolicy),
		RoleTags:             defaultTags(req.OrgId),
	}
	cdrResp, err := execCreateIAMRole(ctx, act, cdrReq)
	if err != nil {
		return fmt.Errorf("unable to create IAM role: %w", err)
	}

	l.Debug("creating iam policy attachment for org %s", req.OrgId)
	cdpaReq := CreateIAMRolePolicyAttachmentRequest{
		OrgsIAMAccessRoleArn: w.cfg.OrgsIAMAccessRoleArn,
		PolicyArn:            cdpResp.PolicyArn,
		RoleArn:              cdrResp.RoleArn,
	}
	err = execCreateIAMRolePolicyAttachment(ctx, act, cdpaReq)
	if err != nil {
		return fmt.Errorf("unable to create IAM role policy attachment: %w", err)
	}

	return nil
}

func (w wkflow) provisionDeploymentsIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) error {
	l := log.With(workflow.GetLogger(ctx))
	act := NewActivities()

	// create deployments
	l.Debug("creating iam policy for org %s", req.OrgId)
	deploymentsPolicy, err := deploymentsIAMPolicy(w.cfg.OrgDeploymentsBucketName, req.OrgId)
	if err != nil {
		return fmt.Errorf("unable to create IAM policy document: %w", err)
	}
	cdpReq := CreateIAMPolicyRequest{
		OrgsIAMAccessRoleArn: w.cfg.OrgsIAMAccessRoleArn,
		PolicyName:           deploymentsIAMName(req.OrgId),
		PolicyPath:           defaultIAMPolicyPath,
		PolicyDocument:       string(deploymentsPolicy),
	}
	cdpResp, err := execCreateIAMPolicy(ctx, act, cdpReq)
	if err != nil {
		return fmt.Errorf("unable to create IAM policy: %w", err)
	}

	l.Debug("creating iam role for org %s", req.OrgId)
	deploymentsTrustPolicy, err := deploymentsIAMTrustPolicy(w.cfg)
	if err != nil {
		return fmt.Errorf("unable to create IAM trust policy document: %w", err)
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
		return fmt.Errorf("unable to create IAM role: %w", err)
	}

	l.Debug("creating iam policy attachment for org %s", req.OrgId)
	cdpaReq := CreateIAMRolePolicyAttachmentRequest{
		OrgsIAMAccessRoleArn: w.cfg.OrgsIAMAccessRoleArn,
		PolicyArn:            cdpResp.PolicyArn,
		RoleArn:              cdrResp.RoleArn,
	}
	err = execCreateIAMRolePolicyAttachment(ctx, act, cdpaReq)
	if err != nil {
		return fmt.Errorf("unable to create IAM role policy attachment: %w", err)
	}

	return nil
}

// ProvisionIAM is a workflow that creates org specific IAM roles in the designated orgs IAM account
func (w wkflow) ProvisionIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (*iamv1.ProvisionIAMResponse, error) {
	resp := &iamv1.ProvisionIAMResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	if err := w.provisionDeploymentsIAM(ctx, req); err != nil {
		return resp, fmt.Errorf("unable to provision deployments IAM role: %w", err)
	}

	if err := w.provisionInstallationsIAM(ctx, req); err != nil {
		return resp, fmt.Errorf("unable to provision installations IAM role: %w", err)
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
) error {
	l := workflow.GetLogger(ctx)

	l.Debug("executing create iam role policy attachment activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateIAMRolePolicyAttachment, req)

	var resp CreateDeploymentsBucketRoleResponse
	if err := fut.Get(ctx, &resp); err != nil {
		return err
	}

	return nil
}
