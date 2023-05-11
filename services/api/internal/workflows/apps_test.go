package workflows

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	tclient "go.temporal.io/sdk/client"
	tmock "go.temporal.io/sdk/mocks"
)

func Test_appWorkflowManager_Provision(t *testing.T) {
	errAppProvisionTest := fmt.Errorf("error")
	app := generics.GetFakeObj[*models.App]()

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

				req, ok := args[0].(*appsv1.ProvisionRequest)
				assert.True(t, ok)

				assert.True(t, ok)
				assert.Equal(t, app.ID, req.AppId)
				assert.Equal(t, app.OrgID, req.OrgId)
			},
		},
		"error": {
			clientFn: func() temporalClient {
				client := &testTemporalClient{}
				client.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errAppProvisionTest)
				return client
			},
			errExpected: errAppProvisionTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.clientFn()
			mgr := NewAppWorkflowManager(client)

			_, err := mgr.Provision(context.Background(), app)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client)
		})
	}
}
