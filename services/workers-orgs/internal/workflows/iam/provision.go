package iam

import (
	"fmt"

	iamv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/iam/v1"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/roles"
	"go.temporal.io/sdk/workflow"
)

func (w wkflow) provisionSecretsIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (string, error) {
	set := iamSet{
		name: "secrets",
		policyFn: func() ([]byte, error) {
			return roles.SecretsIAMPolicy(w.cfg.OrgSecretsBucketName, req.OrgId)
		},
		iamNameFn: func() string {
			return roles.SecretsIAMName(req.OrgId)
		},
		trustPolicyFn: func() ([]byte, error) {
			return roles.InstallerIAMTrustPolicy(w.cfg.WorkersIAMRoleARNPrefix, w.cfg.SupportIAMRoleARN)
		},
	}

	return w.createIAMSet(ctx, req, set)
}

func (w wkflow) provisionOdrIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (string, error) {
	set := iamSet{
		name: "odr",
		policyFn: func() ([]byte, error) {
			return roles.OdrIAMPolicy(w.cfg.OrgsECRRegistryArn, req.OrgId)
		},
		iamNameFn: func() string {
			return roles.OdrIAMName(req.OrgId)
		},
		trustPolicyFn: func() ([]byte, error) {
			return roles.OdrIAMTrustPolicy(w.cfg.OrgsIAMOidcProviderArn, w.cfg.OrgsIAMOidcProviderURL, req.OrgId)
		},
	}
	return w.createIAMSet(ctx, req, set)
}

func (w wkflow) provisionInstallerIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (string, error) {
	set := iamSet{
		name: "installer",
		policyFn: func() ([]byte, error) {
			return roles.InstallerIAMPolicy(w.cfg.OrgInstallationsBucketName, req.OrgId)
		},
		iamNameFn: func() string {
			return roles.InstallerIAMName(req.OrgId)
		},
		trustPolicyFn: func() ([]byte, error) {
			return roles.InstallerIAMTrustPolicy(w.cfg.WorkersIAMRoleARNPrefix, w.cfg.SupportIAMRoleARN)
		},
	}
	return w.createIAMSet(ctx, req, set)
}

func (w wkflow) provisionInstallationsIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (string, error) {
	set := iamSet{
		name: "installations",
		policyFn: func() ([]byte, error) {
			return roles.InstallationsIAMPolicy(
				w.cfg.OrgInstallationsBucketName,
				req.OrgId,
				w.cfg.SandboxBucketARN,
				w.cfg.SandboxKeyARN,
			)
		},
		iamNameFn: func() string {
			return roles.InstallationsIAMName(req.OrgId)
		},
		trustPolicyFn: func() ([]byte, error) {
			return roles.InstallationsIAMTrustPolicy(w.cfg.WorkersIAMRoleARNPrefix, w.cfg.SupportIAMRoleARN)
		},
	}
	return w.createIAMSet(ctx, req, set)
}

func (w wkflow) provisionOrgsIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (string, error) {
	set := iamSet{
		name: "orgs",
		policyFn: func() ([]byte, error) {
			return roles.OrgsIAMPolicy(w.cfg.OrgsBucketName, req.OrgId)
		},
		iamNameFn: func() string {
			return roles.OrgsIAMName(req.OrgId)
		},
		trustPolicyFn: func() ([]byte, error) {
			return roles.InstallationsIAMTrustPolicy(w.cfg.WorkersIAMRoleARNPrefix, w.cfg.SupportIAMRoleARN)
		},
	}
	return w.createIAMSet(ctx, req, set)
}

func (w wkflow) provisionDeploymentsIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (string, error) {
	set := iamSet{
		name: "deployments",
		policyFn: func() ([]byte, error) {
			return roles.DeploymentsIAMPolicy(w.cfg.OrgDeploymentsBucketName, req.OrgId)
		},
		iamNameFn: func() string {
			return roles.DeploymentsIAMName(req.OrgId)
		},
		trustPolicyFn: func() ([]byte, error) {
			return roles.DeploymentsIAMTrustPolicy(w.cfg.WorkersIAMRoleARNPrefix, w.cfg.SupportIAMRoleARN)
		},
	}
	return w.createIAMSet(ctx, req, set)
}

func (w wkflow) provisionInstancesIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (string, error) {
	set := iamSet{
		name: "instances",
		policyFn: func() ([]byte, error) {
			return roles.InstancesIAMPolicy(req.OrgId)
		},
		iamNameFn: func() string {
			return roles.InstancesIAMName(req.OrgId)
		},
		trustPolicyFn: func() ([]byte, error) {
			return roles.InstancesIAMTrustPolicy(w.cfg.WorkersIAMRoleARNPrefix, w.cfg.SupportIAMRoleARN)
		},
	}
	return w.createIAMSet(ctx, req, set)
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

	orgsRoleArn, err := w.provisionOrgsIAM(ctx, req)
	if err != nil {
		return resp, fmt.Errorf("unable to provision orgs IAM role: %w", err)
	}
	resp.OrgsRoleArn = orgsRoleArn

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

	installerRoleArn, err := w.provisionInstallerIAM(ctx, req)
	if err != nil {
		return resp, fmt.Errorf("unable to provision installer IAM role: %w", err)
	}
	resp.InstallerRoleArn = installerRoleArn

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

	secretsRoleArn, err := w.provisionSecretsIAM(ctx, req)
	if err != nil {
		return resp, fmt.Errorf("unable to provision instance IAM role: %w", err)
	}
	resp.SecretsRoleArn = secretsRoleArn
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
