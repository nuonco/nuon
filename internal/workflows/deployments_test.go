package workflows

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/api/internal/models"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_deploymentWorkflowManager_Start(t *testing.T) {
	errDeploymentProvisionTest := fmt.Errorf("error")
	deployment := getFakeObj[*models.Deployment]()
	install := getFakeObj[models.Install]()
	deployment.Component.App.Installs = []models.Install{install}

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
				req, ok := args[0].(*deploymentsv1.StartRequest)
				assert.True(t, ok)

				// make sure the request is valid
				assert.NotNil(t, req)
				assert.NoError(t, req.Validate())

				// make sure all ids are correcctly set
				app := deployment.Component.App
				assert.Equal(t, deployment.ID.String(), req.DeploymentId)
				assert.Equal(t, app.ID.String(), req.AppId)
				assert.Equal(t, app.OrgID.String(), req.OrgId)
				assert.Equal(t, install.ID.String(), req.InstallIds[0])
			},
		},
		"error": {
			clientFn: func() temporalClient {
				client := &testTemporalClient{}
				client.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errDeploymentProvisionTest)
				return client
			},
			errExpected: errDeploymentProvisionTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.clientFn()
			mgr := NewDeploymentWorkflowManager(client)

			err := mgr.Start(context.Background(), deployment)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client)
		})
	}
}
