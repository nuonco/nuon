package workflows

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/generics"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	tclient "go.temporal.io/sdk/client"
	tmock "go.temporal.io/sdk/mocks"
)

func Test_installWorkflowManager_Provision(t *testing.T) {
	errInstallProvisionTest := fmt.Errorf("error")
	install := generics.GetFakeObj[*models.Install]()
	install.AWSSettings = generics.GetFakeObj[*models.AWSSettings]()
	sandboxVersion := generics.GetFakeObj[*models.SandboxVersion]()

	orgID, _ := shortid.NewNanoID("org")
	installID, err := shortid.ParseString(install.ID.String())
	assert.NoError(t, err)

	tests := map[string]struct {
		clientFn    func() temporalClient
		assertFn    func(*testing.T, temporalClient)
		errExpected error
	}{
		"happy path": {
			clientFn: func() temporalClient {
				client := &testTemporalClient{}
				workflowRun := &tmock.WorkflowRun{}
				client.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun, nil)
				workflowRun.On("GetID", mock.Anything, mock.Anything).Return("12345")
				return client
			},
			assertFn: func(t *testing.T, client temporalClient) {
				obj := client.(*testTemporalClient)
				obj.AssertNumberOfCalls(t, "ExecuteWorkflow", 1)

				args, ok := obj.Calls[0].Arguments[3].([]interface{})
				assert.True(t, ok)

				opts, ok := obj.Calls[0].Arguments[1].(tclient.StartWorkflowOptions)
				assert.True(t, ok)
				assert.Equal(t, workflows.DefaultTaskQueue, opts.TaskQueue)

				req, ok := args[0].(*installsv1.ProvisionRequest)
				assert.True(t, ok)

				assert.Equal(t, orgID, req.OrgId)
				assert.Equal(t, installID, req.InstallId)
				assert.Equal(t, install.AppID, req.AppId)
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

			_, err := mgr.Provision(context.Background(), install, orgID, sandboxVersion)
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

	orgID, _ := shortid.NewNanoID("org")
	installID, err := shortid.ParseString(install.ID.String())
	assert.NoError(t, err)

	tests := map[string]struct {
		clientFn    func() temporalClient
		assertFn    func(*testing.T, temporalClient)
		errExpected error
	}{
		"happy path": {
			clientFn: func() temporalClient {
				client := &testTemporalClient{}
				workflowRun := &tmock.WorkflowRun{}
				client.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun, nil)
				workflowRun.On("GetID", mock.Anything, mock.Anything).Return("12345")
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
				assert.Equal(t, install.AppID, req.AppId)
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

			_, err := mgr.Deprovision(context.Background(), install, orgID, sandboxVersion)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client)
		})
	}
}
