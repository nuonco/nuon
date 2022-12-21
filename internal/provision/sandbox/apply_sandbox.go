package sandbox

import (
	"context"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-terraform"
)

const (
	defaultTerraformVersion = "v1.3.6"
	defaultStateFilename    = "state.tf"
)

type AccountSettings struct {
	AwsRegion    string `json:"aws_region" validate:"required"`
	AwsAccountID string `json:"aws_account_id" validate:"required"`
	AwsRoleArn   string `json:"aws_role_arn" validate:"required"`
}

type ApplySandboxRequest struct {
	OrgID               string `json:"org_id" validate:"required"`
	AppID               string `json:"app_id" validate:"required"`
	InstallID           string `json:"install_id" validate:"required"`
	BackendBucketName   string `json:"bucket" validate:"required"`
	BackendBucketRegion string `json:"bucket_region" validate:"required"`
	NuonAccessRoleArn   string `json:"nuon_access_role_arn" validate:"required"`
	OrgAccountID        string `json:"org_account_id" validate:"required"`

	AccountSettings *AccountSettings `json:"account_settings" validate:"required"`

	SandboxSettings struct {
		Name    string `json:"name" validate:"required"`
		Version string `json:"version" validate:"required"`
	} `json:"sandbox_settings"`
	SandboxBucketName string `json:"sandbox_bucket_name" validate:"required"`
}

type ApplySandboxResponse struct {
	Outputs map[string]string
}

func (d ApplySandboxRequest) validate() error {
	validate := validator.New()
	return validate.Struct(d)
}

type terraformApplyer interface {
	provisionSandbox(context.Context, terraformRunnerFn, ApplySandboxRequest) (map[string]string, error)
}

var _ terraformApplyer = (*tfApplyer)(nil)

type tfApplyer struct{}

func getSandboxBucketKey(name, version string) string {
	return fmt.Sprintf("sandboxes/%s_%s.tar.gz", name, version)
}

func getStateBucketKey(orgID, appID, installID string) string {
	return fmt.Sprintf("installations/org=%s/app=%s/install=%s/%s", orgID, appID, installID, defaultStateFilename)
}

func (t *tfApplyer) provisionSandbox(ctx context.Context, fn terraformRunnerFn, req ApplySandboxRequest) (map[string]string, error) {
	runReq := terraform.RunRequest{
		ID:      req.InstallID,
		RunType: terraform.RunTypePlanAndApply,
		Module: terraform.Module{
			BucketName:       req.SandboxBucketName,
			BucketKey:        getSandboxBucketKey(req.SandboxSettings.Name, req.SandboxSettings.Version),
			TerraformVersion: defaultTerraformVersion,
		},
		// TODO(jm): use an s3 logger here, once this works
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		BackendConfig: terraform.BackendConfig{
			BucketRegion: req.BackendBucketRegion,
			BucketName:   req.BackendBucketName,
			BucketKey:    getStateBucketKey(req.OrgID, req.AppID, req.InstallID),
		},
		EnvVars: map[string]string{
			"AWS_REGION": req.AccountSettings.AwsRegion,
		},
		TfVars: map[string]interface{}{
			"nuon_id":             req.InstallID,
			"region":              req.AccountSettings.AwsRegion,
			"assume_role_arn":     req.AccountSettings.AwsRoleArn,
			"install_role_arn":    req.NuonAccessRoleArn,
			"image_sync_role_arn": fmt.Sprintf("arn:aws:iam::%s:role/orgs/org-instances-access-%s", req.OrgAccountID, req.OrgID),
			"tags": map[string]string{
				"nuon_sandbox_name":    req.SandboxSettings.Name,
				"nuon_sandbox_version": req.SandboxSettings.Version,
				"nuon_install_id":      req.InstallID,
				"nuon_app_id":          req.AppID,
			},
		},
	}

	resp, err := fn(ctx, runReq)
	if err != nil {
		return nil, fmt.Errorf("terraform run failed: %w", err)
	}

	return resp.Output, nil
}

// terraformRunnerFn is the client interface for dispatching a terraform run
type terraformRunnerFn func(context.Context, terraform.RunRequest) (terraform.RunResponse, error)

// ApplySandbox: fetches the correct sandbox source terraform, and runs a provision step building the sandbox
func (a *Activities) ApplySandbox(ctx context.Context, req ApplySandboxRequest) (ApplySandboxResponse, error) {
	resp := ApplySandboxResponse{}
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	outputs, err := a.terraformApplyer.provisionSandbox(ctx, terraform.Run, req)
	if err != nil {
		return resp, fmt.Errorf("unable to provision sandbox: %w", err)
	}
	resp.Outputs = outputs

	return resp, err
}
