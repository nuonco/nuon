package client

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/pkg/generics"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/stretchr/testify/assert"
	tmock "go.temporal.io/sdk/mocks"
)

func Test_repo_TriggerInstallDeprovision(t *testing.T) {
	errDeprovision := fmt.Errorf("error deprovision")
	req := generics.GetFakeObj[*installsv1.DeprovisionRequest]()
	tests := map[string]struct {
		client      func(*testing.T, *gomock.Controller) temporal.Client
		errExpected error
	}{
		"happy path": {
			client: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				client := temporal.NewMockClient(mockCtl)
				workflowRun := &tmock.WorkflowRun{}
				workflowRun.On("GetID").Return("12345")
				client.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "installs", gomock.Any(), "Deprovision", gomock.Any()).Return(workflowRun, nil)
				return client
			},
		},
		"error": {
			client: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				client := temporal.NewMockClient(mockCtl)
				client.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "installs", gomock.Any(), "Deprovision", req).Return(nil, errDeprovision)
				return client
			},
			errExpected: errDeprovision,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)

			client := test.client(t, mockCtl)
			repo := &workflowsClient{
				TemporalClient: client,
			}

			_, err := repo.TriggerInstallDeprovision(ctx, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func Test_repo_TriggerInstallProvision(t *testing.T) {
	errProvision := fmt.Errorf("error provision")
	req := generics.GetFakeObj[*installsv1.ProvisionRequest]()
	tests := map[string]struct {
		client      func(*testing.T, *gomock.Controller) temporal.Client
		errExpected error
	}{
		"happy path": {
			client: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				client := temporal.NewMockClient(mockCtl)
				workflowRun := &tmock.WorkflowRun{}
				workflowRun.On("GetID").Return("12345")
				client.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "installs", gomock.Any(), "Provision", gomock.Any()).Return(workflowRun, nil)
				return client
			},
		},
		"error": {
			client: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				client := temporal.NewMockClient(mockCtl)
				client.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "installs", gomock.Any(), "Provision", req).Return(nil, errProvision)
				return client
			},
			errExpected: errProvision,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)

			client := test.client(t, mockCtl)
			repo := &workflowsClient{
				TemporalClient: client,
			}

			_, err := repo.TriggerInstallProvision(ctx, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
