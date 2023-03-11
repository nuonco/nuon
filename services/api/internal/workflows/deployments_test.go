package workflows

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/generics"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/encoding/protojson"
)

func Test_deploymentWorkflowManager_Start(t *testing.T) {
	errDeploymentProvisionTest := fmt.Errorf("error")
	deployment := generics.GetFakeObj[*models.Deployment]()
	// TODO: add valid component config
	component := generics.GetFakeObj[*componentv1.Component]()
	byts, err := protojson.Marshal(component)
	assert.NoError(t, err)
	deployment.Component.Config = byts
	install := generics.GetFakeObj[models.Install]()
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

				assert.Equal(t, shortid.ParseUUID(deployment.ID), req.DeploymentId)
				assert.Equal(t, shortid.ParseUUID(app.ID), req.AppId)
				assert.Equal(t, shortid.ParseUUID(app.OrgID), req.OrgId)

				assert.Equal(t, shortid.ParseUUID(deployment.Component.ID), req.Component.Id)
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
			assert.NoError(t, err)
			test.assertFn(t, client)
		})
	}
}
