package workflows

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/clients/temporal"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
	tmock "go.temporal.io/sdk/mocks"
)

func Test_installWorkflowManager_Provision(t *testing.T) {
	errInstallProvisionTest := fmt.Errorf("error")
	install := generics.GetFakeObj[*models.Install]()
	install.AWSSettings = generics.GetFakeObj[*models.AWSSettings]()
	sandboxVersion := generics.GetFakeObj[*models.SandboxVersion]()

	orgID, _ := shortid.NewNanoID("org")
	install.ID, _ = shortid.NewNanoID("inl")

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

				mock.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "installs", gomock.Any(), gomock.Any(), gomock.Any()).Return(workflowRun, nil)
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

				mock.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "installs", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errInstallProvisionTest)
				return mock
			},
			errExpected: errInstallProvisionTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			client := test.clientFn(mockCtl)
			mgr := NewInstallWorkflowManager(client)

			resp, err := mgr.Provision(context.Background(), install, orgID, sandboxVersion)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client, resp)
		})
	}
}

func Test_installWorkflowManager_Deprovision(t *testing.T) {
	errInstallDeprovisionTest := fmt.Errorf("error")
	install := generics.GetFakeObj[*models.Install]()
	install.AWSSettings = generics.GetFakeObj[*models.AWSSettings]()
	sandboxVersion := generics.GetFakeObj[*models.SandboxVersion]()

	orgID, _ := shortid.NewNanoID("org")
	install.ID, _ = shortid.NewNanoID("inl")

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

				mock.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "installs", gomock.Any(), gomock.Any(), gomock.Any()).Return(workflowRun, nil)
				return mock
			},
			assertFn: func(t *testing.T, client temporal.Client, resp string) {
				assert.Equal(t, "12345", resp)
			},
		},
		"error": {
			clientFn: func(mockCtl *gomock.Controller) temporal.Client {
				mock := temporal.NewMockClient(mockCtl)

				mock.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "installs", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errInstallDeprovisionTest)
				return mock
			},
			errExpected: errInstallDeprovisionTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			client := test.clientFn(mockCtl)
			mgr := NewInstallWorkflowManager(client)

			resp, err := mgr.Deprovision(context.Background(), install, orgID, sandboxVersion)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client, resp)
		})
	}
}
