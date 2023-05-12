package createdeployment

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/workflows"
	"github.com/stretchr/testify/assert"
)

func Test_ActivityTriggerDeploymentJob(t *testing.T) {
	err := errors.New("error")
	deploymentID, _ := shortid.NewNanoID("dpl")
	deployment := generics.GetFakeObj[*models.Deployment]()

	tests := map[string]struct {
		deploymentRepo func(*gomock.Controller) *repos.MockDeploymentRepo
		mockMgr        func(*gomock.Controller) *workflows.MockDeploymentWorkflowManager
		errExpected    error
	}{
		"happy path": {
			deploymentRepo: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				mockRepo := repos.NewMockDeploymentRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), deploymentID).Return(deployment, nil)
				return mockRepo
			},
			mockMgr: func(ctl *gomock.Controller) *workflows.MockDeploymentWorkflowManager {
				wkflowmgr := workflows.NewMockDeploymentWorkflowManager(ctl)
				wkflowmgr.EXPECT().Start(gomock.Any(), deployment).Return("1234", nil)
				return wkflowmgr
			},
		},
		"repo err": {
			deploymentRepo: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				mockRepo := repos.NewMockDeploymentRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), deploymentID).Return(nil, err)
				return mockRepo
			},
			mockMgr: func(ctl *gomock.Controller) *workflows.MockDeploymentWorkflowManager {
				wkflowmgr := workflows.NewMockDeploymentWorkflowManager(ctl)
				return wkflowmgr
			},
			errExpected: err,
		},
		"mgr err": {
			deploymentRepo: func(ctl *gomock.Controller) *repos.MockDeploymentRepo {
				mockRepo := repos.NewMockDeploymentRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), deploymentID).Return(deployment, nil)
				return mockRepo
			},
			mockMgr: func(ctl *gomock.Controller) *workflows.MockDeploymentWorkflowManager {
				wkflowmgr := workflows.NewMockDeploymentWorkflowManager(ctl)
				wkflowmgr.EXPECT().Start(gomock.Any(), deployment).Return("", err)
				return wkflowmgr
			},
			errExpected: err,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			act := &activities{
				repo: test.deploymentRepo(mockCtl),
				mgr:  test.mockMgr(mockCtl),
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
