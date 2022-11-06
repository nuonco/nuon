package deprovision

import (
	"context"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-terraform"
)

const (
	defaultTerraformVersion string = "v1.2.9"
	defaultStateFilename    string = "state.tf"
)

type DestroyTerraformRequest struct {
	DeprovisionRequest `json:"deprovision_request" validate:"required"`

	InstallationStateBucketName   string `json:"installation_state_bucket_name" validate:"required"`
	InstallationStateBucketRegion string `json:"installation_state_bucket_region" validate:"required"`
	SandboxBucketName             string `json:"sandbox_bucket_name" validate:"required"`

	// NuonAssumeRoleArn is the role we add to the k8s cluster to give us access after provisioning
	NuonAssumeRoleArn string `json:"nuon_assume_role_arn" validate:"required"`
}

func (d DestroyTerraformRequest) validate() error {
	validate := validator.New()
	return validate.Struct(d)
}

type DestroyTerraformResponse struct{}

type terraformDestroyer interface {
	destroyTerraform(context.Context, terraformRunnerFn, DestroyTerraformRequest) error
}

var _ terraformDestroyer = (*tfDestroyer)(nil)

type tfDestroyer struct{}

func (t *tfDestroyer) destroyTerraform(ctx context.Context, fn terraformRunnerFn, req DestroyTerraformRequest) error {
	runReq := terraform.RunRequest{
		ID:      req.InstallID,
		RunType: terraform.RunTypeDestroy,
		Module: terraform.Module{
			BucketName:       req.SandboxBucketName,
			BucketKey:        getSandboxBucketKey(req.SandboxSettings.Name, req.SandboxSettings.Version),
			TerraformVersion: defaultTerraformVersion,
		},
		// TODO(jm): use an s3 logger here, once this works
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		BackendConfig: terraform.BackendConfig{
			BucketRegion: req.InstallationStateBucketRegion,
			BucketName:   req.InstallationStateBucketName,
			BucketKey:    getStateBucketKey(req.OrgID, req.AppID, req.InstallID),
		},
		EnvVars: map[string]string{
			"AWS_REGION": req.AwsRegion,
		},
		TfVars: map[string]interface{}{
			"nuon_id":          req.InstallID,
			"region":           req.AwsRegion,
			"assume_role_arn":  req.AssumeRoleArn,
			"install_role_arn": req.NuonAssumeRoleArn,
			"tags": map[string]string{
				"nuon_sandbox_name":    req.SandboxSettings.Name,
				"nuon_sandbox_version": req.SandboxSettings.Version,
				"nuon_install_id":      req.InstallID,
				"nuon_app_id":          req.AppID,
			},
		},
	}

	if _, err := fn(ctx, runReq); err != nil {
		return fmt.Errorf("terraform run failed: %w", err)
	}
	return nil
}

// terraformRunnerFn is the client interface for dispatching a terraform run
type terraformRunnerFn func(context.Context, terraform.RunRequest) (terraform.RunResponse, error)

func (a *Activities) DestroyTerraform(ctx context.Context, req DestroyTerraformRequest) (DestroyTerraformResponse, error) {
	var resp DestroyTerraformResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	err := a.destroyTerraform(ctx, terraform.Run, req)
	return resp, err
}
