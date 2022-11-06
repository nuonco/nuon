package provision

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-terraform"
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

func Test_tfProvisioner_provisionSandbox(t *testing.T) {
	errUnableToRun := fmt.Errorf("unable to run")
	req := getFakeProvisionSandboxRequest()

	tests := map[string]struct {
		tfRunFn     func() *fakeTerraformRunnerFn
		assertFn    func(*testing.T, *fakeTerraformRunnerFn, map[string]string)
		errExpected error
	}{
		"happy path": {
			tfRunFn: func() *fakeTerraformRunnerFn {
				ftr := &fakeTerraformRunnerFn{}
				ftr.On("Run", mock.Anything, mock.Anything).Return(terraform.RunResponse{}, nil)
				return ftr
			},
			assertFn: func(t *testing.T, obj *fakeTerraformRunnerFn, _ map[string]string) {
				obj.AssertNumberOfCalls(t, "Run", 1)
				actualReq := obj.Calls[0].Arguments[1].(terraform.RunRequest)
				assert.NotEmpty(t, actualReq)

				assert.Equal(t, req.InstallID, actualReq.ID)
				assert.Equal(t, terraform.RunTypePlanAndApply, actualReq.RunType)

				// module
				expectedSandboxKey := getSandboxBucketKey(req.SandboxSettings.Name, req.SandboxSettings.Version)
				assert.Equal(t, expectedSandboxKey, actualReq.Module.BucketKey)

				// backend config
				assert.Equal(t, req.BackendBucketName, actualReq.BackendConfig.BucketName)
				assert.Equal(t, req.BackendBucketRegion, actualReq.BackendConfig.BucketRegion)
				expectedBackendKey := getStateBucketKey(req.OrgID, req.AppID, req.InstallID)
				assert.Equal(t, expectedBackendKey, actualReq.BackendConfig.BucketKey)

				// env vars
				assert.Equal(t, req.AccountSettings.AwsRegion, actualReq.EnvVars["AWS_REGION"])

				// tf vars
				assert.Equal(t, req.InstallID, actualReq.TfVars["nuon_id"])
				assert.Equal(t, req.AccountSettings.AwsRegion, actualReq.TfVars["region"])
				assert.Equal(t, req.AccountSettings.AwsRoleArn, actualReq.TfVars["assume_role_arn"])
				assert.Equal(t, req.NuonAccessRoleArn, actualReq.TfVars["install_role_arn"])
			},
			errExpected: nil,
		},
		"returns outputs": {
			tfRunFn: func() *fakeTerraformRunnerFn {
				ftr := &fakeTerraformRunnerFn{}
				ftr.On("Run", mock.Anything, mock.Anything).Return(terraform.RunResponse{
					Output: map[string]string{
						"key": "value",
					},
				}, nil)
				return ftr
			},
			assertFn: func(t *testing.T, obj *fakeTerraformRunnerFn, output map[string]string) {
				assert.Equal(t, output["key"], "value")
			},
			errExpected: nil,
		},
		"error": {
			tfRunFn: func() *fakeTerraformRunnerFn {
				ftr := &fakeTerraformRunnerFn{}
				ftr.On("Run", mock.Anything, mock.Anything).Return(terraform.RunResponse{}, errUnableToRun)
				return ftr
			},
			assertFn: func(t *testing.T, ftr *fakeTerraformRunnerFn, _ map[string]string) {
			},
			errExpected: errUnableToRun,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tr := &tfProvisioner{}

			ftr := test.tfRunFn()
			output, err := tr.provisionSandbox(context.Background(), ftr.Run, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
			} else {
				assert.Nil(t, err)
			}
			test.assertFn(t, ftr, output)
		})
	}
}

func Test_getSandboxBucketKey(t *testing.T) {
	sandboxName := "aws-eks"
	sandboxVersion := "v0.0.1"

	expected := fmt.Sprintf("sandboxes/%s_%s.tar.gz", sandboxName, sandboxVersion)
	assert.Equal(t, expected, getSandboxBucketKey(sandboxName, sandboxVersion))
}

func Test_getStateBucketKey(t *testing.T) {
	orgID := "org123"
	appID := "app123"
	installID := "install123"

	expected := fmt.Sprintf("installations/org=%s/app=%s/install=%s/%s", orgID, appID, installID, defaultStateFilename)
	assert.Equal(t, expected, getStateBucketKey(orgID, appID, installID))
}

func getFakeProvisionSandboxRequest() ProvisionSandboxRequest {
	return ProvisionSandboxRequest{
		InstallID: uuid.New().String(),
		OrgID:     uuid.New().String(),
		AppID:     uuid.New().String(),
		AccountSettings: &AccountSettings{
			AwsAccountID: uuid.New().String(),
			AwsRegion:    validInstallRegions()[0],
			AwsRoleArn:   "arn:aws:something",
		},
		SandboxSettings: struct {
			Name    string `json:"name" validate:"required"`
			Version string `json:"version" validate:"required"`
		}{
			Name:    "aws-eks2",
			Version: "v0.0.1",
		},
	}
}
