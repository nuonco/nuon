package workflows

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/generics"
	installsv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/installs/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_installWorkflowManager_Provision(t *testing.T) {
	errInstallProvisionTest := fmt.Errorf("error")
	install := generics.GetFakeObj[*models.Install]()
	install.AWSSettings = generics.GetFakeObj[*models.AWSSettings]()
	sandboxVersion := generics.GetFakeObj[*models.SandboxVersion]()

	reqOrgID := uuid.NewString()

	shortIDs, err := shortid.ParseStrings(reqOrgID, install.AppID.String(), install.ID.String())
	assert.NoError(t, err)
	orgID, appID, installID := shortIDs[0], shortIDs[1], shortIDs[2]

	tests := map[string]struct {
		clientFn    func() temporalClient
		assertFn    func(*testing.T, temporalClient)
		errExpected error
	}{
		"happy path": {
			clientFn: func() temporalClient {
				client := &testTemporalClient{}
				client.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				return client
			},
			assertFn: func(t *testing.T, client temporalClient) {
				obj := client.(*testTemporalClient)
				obj.AssertNumberOfCalls(t, "ExecuteWorkflow", 1)

				args, ok := obj.Calls[0].Arguments[3].([]interface{})
				assert.True(t, ok)
				req, ok := args[0].(*installsv1.ProvisionRequest)
				assert.True(t, ok)

				assert.Equal(t, orgID, req.OrgId)
				assert.Equal(t, installID, req.InstallId)
				assert.Equal(t, appID, req.AppId)
				// validate account settings
				assert.Equal(t, install.AWSSettings.Region.ToRegion(), req.AccountSettings.Region)
				assert.Equal(t, install.AWSSettings.IamRoleArn, req.AccountSettings.AwsRoleArn)
				// validate sandbox settings
				assert.Equal(t, sandboxVersion.SandboxName, req.SandboxSettings.Name)
				assert.Equal(t, sandboxVersion.SandboxVersion, req.SandboxSettings.Version)
			},
		},
		"error": {
			clientFn: func() temporalClient {
				client := &testTemporalClient{}
				client.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errInstallProvisionTest)
				return client
			},
			errExpected: errInstallProvisionTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.clientFn()
			mgr := NewInstallWorkflowManager(client)

			err := mgr.Provision(context.Background(), install, reqOrgID, sandboxVersion)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client)
		})
	}
}

func Test_installWorkflowManager_Deprovision(t *testing.T) {
	errInstallProvisionTest := fmt.Errorf("error")
	install := generics.GetFakeObj[*models.Install]()
	install.AWSSettings = generics.GetFakeObj[*models.AWSSettings]()
	sandboxVersion := generics.GetFakeObj[*models.SandboxVersion]()

	reqOrgID := uuid.NewString()

	shortIDs, err := shortid.ParseStrings(reqOrgID, install.AppID.String(), install.ID.String())
	assert.NoError(t, err)
	orgID, appID, installID := shortIDs[0], shortIDs[1], shortIDs[2]

	tests := map[string]struct {
		clientFn    func() temporalClient
		assertFn    func(*testing.T, temporalClient)
		errExpected error
	}{
		"happy path": {
			clientFn: func() temporalClient {
				client := &testTemporalClient{}
				client.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				return client
			},
			assertFn: func(t *testing.T, client temporalClient) {
				obj := client.(*testTemporalClient)
				obj.AssertNumberOfCalls(t, "ExecuteWorkflow", 1)

				args, ok := obj.Calls[0].Arguments[3].([]interface{})
				assert.True(t, ok)
				req, ok := args[0].(*installsv1.DeprovisionRequest)
				assert.True(t, ok)

				assert.Equal(t, orgID, req.OrgId)
				assert.Equal(t, installID, req.InstallId)
				assert.Equal(t, appID, req.AppId)
				// validate account settings
				assert.Equal(t, install.AWSSettings.Region.ToRegion(), req.AccountSettings.Region)
				assert.Equal(t, install.AWSSettings.IamRoleArn, req.AccountSettings.AwsRoleArn)
				// validate sandbox settings
				assert.Equal(t, sandboxVersion.SandboxName, req.SandboxSettings.Name)
				assert.Equal(t, sandboxVersion.SandboxVersion, req.SandboxSettings.Version)
			},
		},
		"error": {
			clientFn: func() temporalClient {
				client := &testTemporalClient{}
				client.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errInstallProvisionTest)
				return client
			},
			errExpected: errInstallProvisionTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.clientFn()
			mgr := NewInstallWorkflowManager(client)

			err := mgr.Deprovision(context.Background(), install, reqOrgID, sandboxVersion)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client)
		})
	}
}
