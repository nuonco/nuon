package workflows

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/generics"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
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
	orgID, _ := shortid.NewNanoID("org")
	appID, _ := shortid.NewNanoID("app")
	deploymentID, _ := shortid.NewNanoID("dpl")
	componentID, _ := shortid.NewNanoID("cmp")
	deployment.ID = deploymentID
	deployment.Component.ID = componentID
	deployment.Component.App.ID = appID
	deployment.Component.AppID = appID
	deployment.Component.App.OrgID = orgID

	tests := map[string]struct {
		clientFn    func(*gomock.Controller) workflowsclient.Client
		assertFn    func(*testing.T, workflowsclient.Client, string)
		errExpected error
	}{
		"happy path": {
			clientFn: func(mockCtl *gomock.Controller) workflowsclient.Client {
				mock := workflowsclient.NewMockClient(mockCtl)
				mock.EXPECT().TriggerDeploymentStart(gomock.Any(), gomock.Any()).Return("12345", nil)
				return mock
			},
			assertFn: func(t *testing.T, client workflowsclient.Client, resp string) {
				assert.Equal(t, "12345", resp)
			},
		},
		"error": {
			clientFn: func(mockCtl *gomock.Controller) workflowsclient.Client {
				mock := workflowsclient.NewMockClient(mockCtl)
				mock.EXPECT().TriggerDeploymentStart(gomock.Any(), gomock.Any()).Return("", errDeploymentProvisionTest)
				return mock
			},
			errExpected: errDeploymentProvisionTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			client := test.clientFn(mockCtl)

			mgr := NewDeploymentWorkflowManager(client)

			resp, err := mgr.Start(context.Background(), deployment)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, client, resp)
		})
	}
}
