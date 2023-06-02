package createdeployment

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/generics"
	wfc "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func Test_ActivityTriggerDeploymentJob(t *testing.T) {
	err := errors.New("error")
	deploymentID, _ := shortid.NewNanoID("dpl")
	deployment := generics.GetFakeObj[*models.Deployment]()
	deployment.Component.Config = datatypes.JSON(`{"deployCfg": {}}`)
	tests := map[string]struct {
		deploymentRepo func(*gomock.Controller) *repos.MockDeploymentRepo
		mockWfc        func(*gomock.Controller) wfc.Client
		errExpected    error
	}{
		"happy path": {
			deploymentRepo: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				mockRepo := repos.NewMockDeploymentRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), deploymentID).Return(deployment, nil)
				return mockRepo
			},
			mockWfc: func(ctl *gomock.Controller) wfc.Client {
				wkflowmgr := wfc.NewMockClient(ctl)
				wkflowmgr.EXPECT().TriggerDeploymentStart(gomock.Any(), gomock.Any()).Return("1234", nil)
				return wkflowmgr
			},
		},
		"repo err": {
			deploymentRepo: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				mockRepo := repos.NewMockDeploymentRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), deploymentID).Return(nil, err)
				return mockRepo
			},
			mockWfc: func(ctl *gomock.Controller) wfc.Client {
				mockWfc := wfc.NewMockClient(ctl)
				return mockWfc
			},
			errExpected: err,
		},
		"mgr err": {
			deploymentRepo: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				mockRepo := repos.NewMockDeploymentRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), deploymentID).Return(deployment, nil)
				return mockRepo
			},
			mockWfc: func(ctl *gomock.Controller) wfc.Client {
				mockWfc := wfc.NewMockClient(ctl)
				mockWfc.EXPECT().TriggerDeploymentStart(gomock.Any(), gomock.Any()).Return("", err)
				return mockWfc
			},
			errExpected: err,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			act := &activities{
				repo: test.deploymentRepo(mockCtl),
				wfc:  test.mockWfc(mockCtl),
			}

			_, err := act.TriggerDeploymentJob(context.Background(), deploymentID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
