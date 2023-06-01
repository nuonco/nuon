package workflows

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
)

func Test_appWorkflowManager_Provision(t *testing.T) {
	errAppProvisionTest := fmt.Errorf("error")
	app := generics.GetFakeObj[*models.App]()

	tests := map[string]struct {
		clientFn    func(*gomock.Controller) workflows.Client
		assertFn    func(*testing.T, string)
		errExpected error
	}{
		"happy path": {
			clientFn: func(mockCtl *gomock.Controller) workflows.Client {
				mock := workflows.NewMockClient(mockCtl)
				mock.EXPECT().TriggerAppProvision(gomock.Any(), gomock.Any()).Return("12345", nil)
				return mock
			},
			assertFn: func(_ *testing.T, resp string) {
				// TODO(jm): find a better way to grab captured arguments with mockgen mocks.
				assert.Equal(t, resp, "12345")
			},
		},
		"error": {
			clientFn: func(mockCtl *gomock.Controller) workflows.Client {
				mock := workflows.NewMockClient(mockCtl)
				mock.EXPECT().TriggerAppProvision(gomock.Any(), gomock.Any()).Return("", errAppProvisionTest)
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

			test.assertFn(t, resp)
		})
	}
}
