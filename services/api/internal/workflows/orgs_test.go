package workflows

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/stretchr/testify/assert"
)

func Test_orgWorkflowManager_Provision(t *testing.T) {
	errOrgProvisionTest := fmt.Errorf("error")
	orgID, _ := shortid.NewNanoID("org")

	tests := map[string]struct {
		clientFn    func(*gomock.Controller) workflowsclient.Client
		assertFn    func(*testing.T, workflowsclient.Client, string)
		errExpected error
	}{
		"happy path": {
			clientFn: func(mockCtl *gomock.Controller) workflowsclient.Client {
				mock := workflowsclient.NewMockClient(mockCtl)
				mock.EXPECT().TriggerOrgSignup(gomock.Any(), gomock.Any()).Return("12345", nil)
				return mock
			},
			assertFn: func(t *testing.T, client workflowsclient.Client, resp string) {
				assert.Equal(t, resp, "12345")
			},
		},
		"error": {
			clientFn: func(mockCtl *gomock.Controller) workflowsclient.Client {
				mock := workflowsclient.NewMockClient(mockCtl)
				mock.EXPECT().TriggerOrgSignup(gomock.Any(), gomock.Any()).Return("", errOrgProvisionTest)
				return mock
			},
			errExpected: errOrgProvisionTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			client := test.clientFn(mockCtl)

			mgr := NewOrgWorkflowManager(client)

			resp, err := mgr.Provision(context.Background(), orgID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client, resp)
		})
	}
}

func Test_orgWorkflowManager_Deprovision(t *testing.T) {
	errOrgDeprovisionTest := fmt.Errorf("error")
	orgID, _ := shortid.NewNanoID("org")

	tests := map[string]struct {
		clientFn    func(*gomock.Controller) workflowsclient.Client
		assertFn    func(*testing.T, workflowsclient.Client, string)
		errExpected error
	}{
		"happy path": {
			clientFn: func(mockCtl *gomock.Controller) workflowsclient.Client {
				mock := workflowsclient.NewMockClient(mockCtl)
				mock.EXPECT().TriggerOrgTeardown(gomock.Any(), gomock.Any()).Return("12345", nil)
				return mock
			},
			assertFn: func(t *testing.T, client workflowsclient.Client, resp string) {
				assert.Equal(t, "12345", resp)
			},
		},
		"error": {
			clientFn: func(mockCtl *gomock.Controller) workflowsclient.Client {
				mock := workflowsclient.NewMockClient(mockCtl)
				mock.EXPECT().TriggerOrgTeardown(gomock.Any(), gomock.Any()).Return("", errOrgDeprovisionTest)
				return mock
			},
			errExpected: errOrgDeprovisionTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			client := test.clientFn(mockCtl)

			mgr := NewOrgWorkflowManager(client)

			resp, err := mgr.Deprovision(context.Background(), orgID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client, resp)
		})
	}
}
