package deprovision

import (
	"context"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-terraform"
	installsv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1"
)

const (
	defaultTerraformVersion string = "v1.3.6"
	defaultStateFilename    string = "state.tf"
)

type DestroyTerraformRequest struct {
	DeprovisionRequest *installsv1.DeprovisionRequest `json:"deprovision_request" validate:"required"`

	InstallationsBucketName   string `json:"installations_bucket_name" validate:"required"`
	InstallationsBucketRegion string `json:"installations_bucket_region" validate:"required"`
	SandboxBucketName         string `json:"sandbox_bucket_name" validate:"required"`

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
	dr := req.DeprovisionRequest

	runReq := terraform.RunRequest{
		ID:      dr.InstallId,
		RunType: terraform.RunTypeDestroy,
		Module: terraform.Module{
			BucketName:       req.SandboxBucketName,
			BucketKey:        getSandboxBucketKey(dr.SandboxSettings.Name, dr.SandboxSettings.Version),
			TerraformVersion: defaultTerraformVersion,
		},
		// TODO(jm): use an s3 logger here, once this works
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		BackendConfig: terraform.BackendConfig{
			BucketRegion: req.InstallationsBucketRegion,
			BucketName:   req.InstallationsBucketName,
			BucketKey:    getStateBucketKey(dr.OrgId, dr.AppId, dr.InstallId),
		},
		EnvVars: map[string]string{
			"AWS_REGION": dr.AccountSettings.Region,
		},
		TfVars: map[string]interface{}{
			"nuon_id":                           dr.InstallId,
			"region":                            dr.AccountSettings.Region,
			"assume_role_arn":                   dr.AccountSettings.AwsRoleArn,
			"install_role_arn":                  req.NuonAssumeRoleArn,
			"waypoint_odr_namespace":            dr.InstallId,
			"waypoint_odr_service_account_name": fmt.Sprintf("waypoint-odr-%s", dr.InstallId),
			"tags": map[string]string{
				"nuon_sandbox_name":    dr.SandboxSettings.Name,
				"nuon_sandbox_version": dr.SandboxSettings.Version,
				"nuon_install_id":      dr.InstallId,
				"nuon_app_id":          dr.AppId,
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
