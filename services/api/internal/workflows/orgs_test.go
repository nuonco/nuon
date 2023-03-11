package workflows

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	orgsv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/orgs/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_orgWorkflowManager_Provision(t *testing.T) {
	errOrgProvisionTest := fmt.Errorf("error")
	reqOrgID := uuid.New()
	orgID := shortid.ParseUUID(reqOrgID)

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

			err := mgr.Provision(context.Background(), reqOrgID.String())
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
	reqOrgID := uuid.New()
	orgID := shortid.ParseUUID(reqOrgID)

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

			err := mgr.Deprovision(context.Background(), reqOrgID.String())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client)
		})
	}
}
