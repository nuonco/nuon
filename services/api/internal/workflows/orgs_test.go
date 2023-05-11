package workflows

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	tclient "go.temporal.io/sdk/client"
	tmock "go.temporal.io/sdk/mocks"
)

func Test_orgWorkflowManager_Provision(t *testing.T) {
	errOrgProvisionTest := fmt.Errorf("error")
	orgID, _ := shortid.NewNanoID("org")

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

				req, ok := args[0].(*orgsv1.SignupRequest)
				assert.True(t, ok)

				assert.Equal(t, orgID, req.OrgId)
			},
		},
		"error": {
			clientFn: func() temporalClient {
				client := &testTemporalClient{}
				client.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errOrgProvisionTest)
				return client
			},
			errExpected: errOrgProvisionTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.clientFn()
			mgr := NewOrgWorkflowManager(client)

			_, err := mgr.Provision(context.Background(), orgID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client)
		})
	}
}

func Test_orgWorkflowManager_Deprovision(t *testing.T) {
	errOrgDeprovisionTest := fmt.Errorf("error")
	orgID, _ := shortid.NewNanoID("org")

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

				req, ok := args[0].(orgDeprovisionArgs)
				assert.True(t, ok)

				assert.Equal(t, orgID, req.OrgID)
			},
		},
		"error": {
			clientFn: func() temporalClient {
				client := &testTemporalClient{}
				client.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errOrgDeprovisionTest)
				return client
			},
			errExpected: errOrgDeprovisionTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.clientFn()
			mgr := NewOrgWorkflowManager(client)

			_, err := mgr.Deprovision(context.Background(), orgID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client)
		})
	}
}
