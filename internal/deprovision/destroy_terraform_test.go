package deprovision

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/go-generics"
	"github.com/powertoolsdev/go-terraform"
	installsv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type fakeTerraformRunnerFn struct {
	mock.Mock
}

func (f *fakeTerraformRunnerFn) Run(ctx context.Context, req terraform.RunRequest) (terraform.RunResponse, error) {
	resp := f.Called(ctx, req)
	if resp.Get(0) != nil {
		return resp.Get(0).(terraform.RunResponse), resp.Error(1)
	}

	return terraform.RunResponse{}, resp.Error(1)
}

func Test_tfDestroyer_destroyTerraform(t *testing.T) {
	errUnableToRun := fmt.Errorf("unable to run")
	req := DestroyTerraformRequest{
		DeprovisionRequest:        generics.GetFakeObj[*installsv1.DeprovisionRequest](),
		InstallationsBucketName:   "s3://nuon-installations",
		InstallationsBucketRegion: "aws-west-2",
		SandboxBucketName:         "s3://nuon-sandboxes",
		NuonAssumeRoleArn:         "arn:abc/installer",
	}

	tests := map[string]struct {
		tfRunFn     func() *fakeTerraformRunnerFn
		assertFn    func(*testing.T, *fakeTerraformRunnerFn)
		errExpected error
	}{
		"happy path": {
			tfRunFn: func() *fakeTerraformRunnerFn {
				ftr := &fakeTerraformRunnerFn{}
				ftr.On("Run", mock.Anything, mock.Anything).Return(terraform.RunResponse{}, nil)
				return ftr
			},
			assertFn: func(t *testing.T, obj *fakeTerraformRunnerFn) {
				obj.AssertNumberOfCalls(t, "Run", 1)
				actualReq := obj.Calls[0].Arguments[1].(terraform.RunRequest)
				depReq := req.DeprovisionRequest
				assert.NotEmpty(t, actualReq)

				assert.Equal(t, depReq.InstallId, actualReq.ID)
				assert.Equal(t, terraform.RunTypeDestroy, actualReq.RunType)

				// module
				assert.Equal(t, req.SandboxBucketName, actualReq.Module.BucketName)
				expectedSandboxKey := getSandboxBucketKey(depReq.SandboxSettings.Name, depReq.SandboxSettings.Version)
				assert.Equal(t, expectedSandboxKey, actualReq.Module.BucketKey)

				// backend config
				assert.Equal(t, req.InstallationsBucketName, actualReq.BackendConfig.BucketName)
				assert.Equal(t, req.InstallationsBucketRegion, actualReq.BackendConfig.BucketRegion)
				expectedBackendKey := getStateBucketKey(depReq.OrgId, depReq.AppId, depReq.InstallId)
				assert.Equal(t, expectedBackendKey, actualReq.BackendConfig.BucketKey)

				// env vars
				assert.Equal(t, depReq.AccountSettings.Region, actualReq.EnvVars["AWS_REGION"])

				// tf vars
				assert.Equal(t, depReq.InstallId, actualReq.TfVars["nuon_id"])
				assert.Equal(t, depReq.AccountSettings.Region, actualReq.TfVars["region"])
				assert.Equal(t, depReq.AccountSettings.AwsRoleArn, actualReq.TfVars["assume_role_arn"])
				assert.Equal(t, req.NuonAssumeRoleArn, actualReq.TfVars["install_role_arn"])
			},
			errExpected: nil,
		},
		"error": {
			tfRunFn: func() *fakeTerraformRunnerFn {
				ftr := &fakeTerraformRunnerFn{}
				ftr.On("Run", mock.Anything, mock.Anything).Return(terraform.RunResponse{}, errUnableToRun)
				return ftr
			},
			assertFn: func(t *testing.T, ftr *fakeTerraformRunnerFn) {

			},
			errExpected: errUnableToRun,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tr := &tfDestroyer{}

			ftr := test.tfRunFn()
			err := tr.destroyTerraform(context.Background(), ftr.Run, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
			} else {
				assert.Nil(t, err)
			}
			test.assertFn(t, ftr)
		})
	}
}
