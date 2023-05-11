package createapp

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

func Test_ActivityTriggerAppJob(t *testing.T) {
	err := errors.New("error")
	appID, _ := shortid.NewNanoID("app")
	app := generics.GetFakeObj[*models.App]()

	tests := map[string]struct {
		appRepo     func(*gomock.Controller) *repos.MockAppRepo
		mockMgr     func(*gomock.Controller) *workflows.MockAppWorkflowManager
		errExpected error
	}{
		"happy path": {
			appRepo: func(ctl *gomock.Controller) *repos.MockAppRepo {
				mockRepo := repos.NewMockAppRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), appID).Return(app, nil)
				return mockRepo
			},
			mockMgr: func(ctl *gomock.Controller) *workflows.MockAppWorkflowManager {
				wkflowmgr := workflows.NewMockAppWorkflowManager(ctl)
				wkflowmgr.EXPECT().Provision(gomock.Any(), app).Return("123456", nil)
				return wkflowmgr
			},
		},
		"repo err": {
			appRepo: func(ctl *gomock.Controller) *repos.MockAppRepo {
				mockRepo := repos.NewMockAppRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), appID).Return(nil, err)
				return mockRepo
			},
			mockMgr: func(ctl *gomock.Controller) *workflows.MockAppWorkflowManager {
				wkflowmgr := workflows.NewMockAppWorkflowManager(ctl)
				return wkflowmgr
			},
			errExpected: err,
		},
		"mgr err": {
			appRepo: func(ctl *gomock.Controller) *repos.MockAppRepo {
				mockRepo := repos.NewMockAppRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), appID).Return(app, nil)
				return mockRepo
			},
			mockMgr: func(ctl *gomock.Controller) *workflows.MockAppWorkflowManager {
				wkflowmgr := workflows.NewMockAppWorkflowManager(ctl)
				wkflowmgr.EXPECT().Provision(gomock.Any(), app).Return("", err)
				return wkflowmgr
			},
			errExpected: err,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			act := &activities{
				repo: test.appRepo(mockCtl),
				mgr:  test.mockMgr(mockCtl),
			}

			_, err := act.TriggerAppJob(context.Background(), appID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
