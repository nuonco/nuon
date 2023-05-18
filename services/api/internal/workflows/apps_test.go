package workflows

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/clients/temporal"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
	tmock "go.temporal.io/sdk/mocks"
)

func Test_appWorkflowManager_Provision(t *testing.T) {
	errAppProvisionTest := fmt.Errorf("error")
	app := generics.GetFakeObj[*models.App]()

	tests := map[string]struct {
		clientFn    func(*gomock.Controller) temporal.Client
		assertFn    func(*testing.T, temporal.Client, string)
		errExpected error
	}{
		"happy path": {
			clientFn: func(mockCtl *gomock.Controller) temporal.Client {
				mock := temporal.NewMockClient(mockCtl)

				workflowRun := &tmock.WorkflowRun{}
				workflowRun.On("GetID").Return("12345")

				mock.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "apps", gomock.Any(), gomock.Any(), gomock.Any()).Return(workflowRun, nil)
				return mock
			},
			assertFn: func(_ *testing.T, _ temporal.Client, resp string) {
				// TODO(jm): find a better way to grab captured arguments with mockgen mocks.
				assert.Equal(t, resp, "12345")
			},
		},
		"error": {
			clientFn: func(mockCtl *gomock.Controller) temporal.Client {
				mock := temporal.NewMockClient(mockCtl)

				workflowRun := &tmock.WorkflowRun{}
				workflowRun.On("GetID", gomock.Any(), gomock.Any()).Return("12345")

				mock.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "apps", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errAppProvisionTest)
				return mock
			},
			errExpected: errAppProvisionTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			client := test.clientFn(mockCtl)

			mgr := NewAppWorkflowManager(client)
			resp, err := mgr.Provision(context.Background(), app)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client, resp)
		})
	}
}
