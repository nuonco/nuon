package iam

import (
	"fmt"
	"time"

	iamv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1/iam/v1"
	workers "github.com/powertoolsdev/workers-orgs/internal"
	"github.com/powertoolsdev/workers-orgs/internal/roles"
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

func (w wkflow) provisionOdrIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (string, error) {
	l := log.With(workflow.GetLogger(ctx))
	act := NewActivities()

	l.Debug("creating odr iam policy for org %s", req.OrgId)
	odrPolicy, err := roles.OdrIAMPolicy(w.cfg.OrgsECRRegistryArn, req.OrgId)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM policy document: %w", err)
	}
	cdpReq := CreateIAMPolicyRequest{
		AssumeRoleARN:  w.cfg.OrgsIAMAccessRoleArn,
		PolicyName:     roles.OdrIAMName(req.OrgId),
		PolicyPath:     defaultIAMPolicyPath,
		PolicyDocument: string(odrPolicy),
		PolicyTags:     roles.DefaultTags(req.OrgId),
	}
	cdpResp, err := execCreateIAMPolicy(ctx, act, cdpReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM policy: %w", err)
	}

	l.Debug("creating iam role for org %s", req.OrgId)
	odrTrustPolicy, err := roles.OdrIAMTrustPolicy(w.cfg.OrgsIAMOidcProviderArn, w.cfg.OrgsIAMOidcProviderURL, req.OrgId)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM trust policy document: %w", err)
	}
	cdrReq := CreateIAMRoleRequest{
		AssumeRoleARN:       w.cfg.OrgsIAMAccessRoleArn,
		RoleName:            roles.OdrIAMName(req.OrgId),
		RolePath:            defaultIAMRolePath,
		TrustPolicyDocument: string(odrTrustPolicy),
		RoleTags:            roles.DefaultTags(req.OrgId),
	}
	cdrResp, err := execCreateIAMRole(ctx, act, cdrReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role: %w", err)
	}

	l.Debug("creating iam policy attachment for org %s", req.OrgId)
	cdpaReq := CreateIAMRolePolicyAttachmentRequest{
		AssumeRoleARN: w.cfg.OrgsIAMAccessRoleArn,
		PolicyArn:     cdpResp.PolicyArn,
		RoleName:      roles.OdrIAMName(req.OrgId),
	}
	err = execCreateIAMRolePolicyAttachment(ctx, act, cdpaReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role policy attachment: %w", err)
	}

	return cdrResp.RoleArn, nil
}

func (w wkflow) provisionInstallationsIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (string, error) {
	l := log.With(workflow.GetLogger(ctx))
	act := NewActivities()

	// create deployments
	l.Debug("creating installations iam policy for org %s", req.OrgId)
	installationsPolicy, err := roles.InstallationsIAMPolicy(w.cfg.OrgInstallationsBucketName, req.OrgId)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM policy document: %w", err)
	}
	cdpReq := CreateIAMPolicyRequest{
		AssumeRoleARN:  w.cfg.OrgsIAMAccessRoleArn,
		PolicyName:     roles.InstallationsIAMName(req.OrgId),
		PolicyPath:     defaultIAMPolicyPath,
		PolicyDocument: string(installationsPolicy),
		PolicyTags:     roles.DefaultTags(req.OrgId),
	}
	cdpResp, err := execCreateIAMPolicy(ctx, act, cdpReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM policy: %w", err)
	}

	l.Debug("creating installations iam role for org %s", req.OrgId)
	installationsTrustPolicy, err := roles.InstallationsIAMTrustPolicy(w.cfg.WorkersIAMRoleARNPrefix)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM trust policy document: %w", err)
	}
	cdrReq := CreateIAMRoleRequest{
		AssumeRoleARN:       w.cfg.OrgsIAMAccessRoleArn,
		RoleName:            roles.InstallationsIAMName(req.OrgId),
		RolePath:            defaultIAMRolePath,
		TrustPolicyDocument: string(installationsTrustPolicy),
		RoleTags:            roles.DefaultTags(req.OrgId),
	}
	cdrResp, err := execCreateIAMRole(ctx, act, cdrReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role: %w", err)
	}

	l.Debug("creating installations iam policy attachment for org %s", req.OrgId)
	cdpaReq := CreateIAMRolePolicyAttachmentRequest{
		AssumeRoleARN: w.cfg.OrgsIAMAccessRoleArn,
		PolicyArn:     cdpResp.PolicyArn,
		RoleName:      roles.InstallationsIAMName(req.OrgId),
	}
	err = execCreateIAMRolePolicyAttachment(ctx, act, cdpaReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role policy attachment: %w", err)
	}

	return cdrResp.RoleArn, nil
}

func (w wkflow) provisionDeploymentsIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (string, error) {
	l := log.With(workflow.GetLogger(ctx))
	act := NewActivities()

	// create deployments
	l.Debug("creating deployments iam policy for org %s", req.OrgId)
	deploymentsPolicy, err := roles.DeploymentsIAMPolicy(w.cfg.OrgDeploymentsBucketName, req.OrgId)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM policy document: %w", err)
	}
	cdpReq := CreateIAMPolicyRequest{
		AssumeRoleARN:  w.cfg.OrgsIAMAccessRoleArn,
		PolicyName:     roles.DeploymentsIAMName(req.OrgId),
		PolicyPath:     defaultIAMPolicyPath,
		PolicyDocument: string(deploymentsPolicy),
		PolicyTags:     roles.DefaultTags(req.OrgId),
	}
	cdpResp, err := execCreateIAMPolicy(ctx, act, cdpReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM policy: %w", err)
	}

	l.Debug("creating deployments iam role for org %s", req.OrgId)
	deploymentsTrustPolicy, err := roles.DeploymentsIAMTrustPolicy(w.cfg.WorkersIAMRoleARNPrefix)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM trust policy document: %w", err)
	}
	cdrReq := CreateIAMRoleRequest{
		AssumeRoleARN:       w.cfg.OrgsIAMAccessRoleArn,
		RoleName:            roles.DeploymentsIAMName(req.OrgId),
		RolePath:            defaultIAMRolePath,
		TrustPolicyDocument: string(deploymentsTrustPolicy),
		RoleTags:            roles.DefaultTags(req.OrgId),
	}
	cdrResp, err := execCreateIAMRole(ctx, act, cdrReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role: %w", err)
	}

	l.Debug("creating deployments iam policy attachment for org %s", req.OrgId)
	cdpaReq := CreateIAMRolePolicyAttachmentRequest{
		AssumeRoleARN: w.cfg.OrgsIAMAccessRoleArn,
		PolicyArn:     cdpResp.PolicyArn,
		RoleName:      roles.DeploymentsIAMName(req.OrgId),
	}
	err = execCreateIAMRolePolicyAttachment(ctx, act, cdpaReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role policy attachment: %w", err)
	}

	return cdrResp.RoleArn, nil
}

func (w wkflow) provisionInstancesIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (string, error) {
	l := log.With(workflow.GetLogger(ctx))
	act := NewActivities()

	// create deployments
	l.Debug("creating instances iam policy for org %s", req.OrgId)
	installationsPolicy, err := roles.InstancesIAMPolicy(req.OrgId)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM policy document: %w", err)
	}
	cdpReq := CreateIAMPolicyRequest{
		AssumeRoleARN:  w.cfg.OrgsIAMAccessRoleArn,
		PolicyName:     roles.InstancesIAMName(req.OrgId),
		PolicyPath:     defaultIAMPolicyPath,
		PolicyDocument: string(installationsPolicy),
		PolicyTags:     roles.DefaultTags(req.OrgId),
	}
	cdpResp, err := execCreateIAMPolicy(ctx, act, cdpReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM policy: %w", err)
	}

	l.Debug("creating instances iam role for org %s", req.OrgId)
	installationsTrustPolicy, err := roles.InstancesIAMTrustPolicy(w.cfg.WorkersIAMRoleARNPrefix)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM trust policy document: %w", err)
	}
	cdrReq := CreateIAMRoleRequest{
		AssumeRoleARN:       w.cfg.OrgsIAMAccessRoleArn,
		RoleName:            roles.InstancesIAMName(req.OrgId),
		RolePath:            defaultIAMRolePath,
		TrustPolicyDocument: string(installationsTrustPolicy),
		RoleTags:            roles.DefaultTags(req.OrgId),
	}
	cdrResp, err := execCreateIAMRole(ctx, act, cdrReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role: %w", err)
	}

	l.Debug("creating instances iam policy attachment for org %s", req.OrgId)
	cdpaReq := CreateIAMRolePolicyAttachmentRequest{
		AssumeRoleARN: w.cfg.OrgsIAMAccessRoleArn,
		PolicyArn:     cdpResp.PolicyArn,
		RoleName:      roles.InstancesIAMName(req.OrgId),
	}
	err = execCreateIAMRolePolicyAttachment(ctx, act, cdpaReq)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role policy attachment: %w", err)
	}

	return cdrResp.RoleArn, nil
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

	deploymentsRoleArn, err := w.provisionDeploymentsIAM(ctx, req)
	if err != nil {
		return resp, fmt.Errorf("unable to provision deployments IAM role: %w", err)
	}
	resp.DeploymentsRoleArn = deploymentsRoleArn

	installationsRoleArn, err := w.provisionInstallationsIAM(ctx, req)
	if err != nil {
		return resp, fmt.Errorf("unable to provision installations IAM role: %w", err)
	}
	resp.InstallationsRoleArn = installationsRoleArn

	odrRoleArn, err := w.provisionOdrIAM(ctx, req)
	if err != nil {
		return resp, fmt.Errorf("unable to provision odr IAM role: %w", err)
	}
	resp.OdrRoleArn = odrRoleArn

	instanceRoleArn, err := w.provisionInstancesIAM(ctx, req)
	if err != nil {
		return resp, fmt.Errorf("unable to provision instance IAM role: %w", err)
	}
	resp.InstancesRoleArn = instanceRoleArn

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

	var resp CreateIAMRolePolicyAttachmentResponse
	if err := fut.Get(ctx, &resp); err != nil {
		return err
	}

	return nil
}
