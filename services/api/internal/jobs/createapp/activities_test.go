package createapp

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	"github.com/powertoolsdev/mono/pkg/generics"
	wfc "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/stretchr/testify/assert"
)

func Test_ActivityTriggerAppJob(t *testing.T) {
	err := errors.New("error")
	appID := domains.NewAppID()
	app := generics.GetFakeObj[*models.App]()

	tests := map[string]struct {
		appRepo     func(*gomock.Controller) *repos.MockAppRepo
		mockWfc     func(*gomock.Controller) wfc.Client
		errExpected error
	}{
		"happy path": {
			appRepo: func(ctl *gomock.Controller) *repos.MockAppRepo {
				mockRepo := repos.NewMockAppRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), appID).Return(app, nil)
				return mockRepo
			},
			mockWfc: func(ctl *gomock.Controller) wfc.Client {
				mockWfc := wfc.NewMockClient(ctl)
				mockWfc.EXPECT().TriggerAppProvision(gomock.Any(), app.ToProvisionRequest()).Return("123456", nil)
				return mockWfc
			},
		},
		"repo err": {
			appRepo: func(ctl *gomock.Controller) *repos.MockAppRepo {
				mockRepo := repos.NewMockAppRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), appID).Return(nil, err)
				return mockRepo
			},
			mockWfc: func(ctl *gomock.Controller) wfc.Client {
				return wfc.NewMockClient(ctl)
			},
			errExpected: err,
		},
		"workflow err": {
			appRepo: func(ctl *gomock.Controller) *repos.MockAppRepo {
				mockRepo := repos.NewMockAppRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), appID).Return(app, nil)
				return mockRepo
			},
			mockWfc: func(ctl *gomock.Controller) wfc.Client {
				mockWfc := wfc.NewMockClient(ctl)
				mockWfc.EXPECT().TriggerAppProvision(gomock.Any(), app.ToProvisionRequest()).Return("", err)
				return mockWfc
			},
			errExpected: err,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			act := &activities{
				repo: test.appRepo(mockCtl),
				wfc:  test.mockWfc(mockCtl),
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
