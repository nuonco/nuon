package createinstall

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/generics"
	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/stretchr/testify/assert"
)

func Test_ActivityTriggerInstallJob(t *testing.T) {
	err := errors.New("error")
	app := generics.GetFakeObj[*models.App]()

	installID, _ := shortid.NewNanoID("inl")
	install := generics.GetFakeObj[*models.Install]()
	install.AWSSettings = &models.AWSSettings{Region: models.AWSRegionUsEast1}
	sandboxVersion := generics.GetFakeObj[*models.SandboxVersion]()

	tests := map[string]struct {
		installRepo func(*gomock.Controller) *repos.MockInstallRepo
		appRepo     func(*gomock.Controller) *repos.MockAppRepo
		adminRepo   func(*gomock.Controller) *repos.MockAdminRepo
		mockWfc     func(*gomock.Controller) workflowsclient.Client
		errExpected error
	}{
		"happy path": {
			installRepo: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				mockRepo := repos.NewMockInstallRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), installID).Return(install, nil)
				return mockRepo
			},
			appRepo: func(ctl *gomock.Controller) *repos.MockAppRepo {
				mockRepo := repos.NewMockAppRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), install.AppID).Return(app, nil)
				return mockRepo
			},
			adminRepo: func(ctl *gomock.Controller) *repos.MockAdminRepo {
				mockRepo := repos.NewMockAdminRepo(ctl)
				mockRepo.EXPECT().GetLatestSandboxVersion(gomock.Any()).Return(sandboxVersion, nil)
				return mockRepo
			},
			mockWfc: func(ctl *gomock.Controller) workflowsclient.Client {
				mockWfc := workflowsclient.NewMockClient(ctl)
				mockWfc.EXPECT().TriggerInstallProvision(gomock.Any(), gomock.Any()).Return("1234", nil)
				return mockWfc
			},
		},
		"repo error": {
			installRepo: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				mockRepo := repos.NewMockInstallRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), installID).Return(nil, err)
				return mockRepo
			},
			appRepo: func(ctl *gomock.Controller) *repos.MockAppRepo {
				mockRepo := repos.NewMockAppRepo(ctl)
				return mockRepo
			},
			adminRepo: func(ctl *gomock.Controller) *repos.MockAdminRepo {
				mockRepo := repos.NewMockAdminRepo(ctl)
				return mockRepo
			},
			mockWfc: func(ctl *gomock.Controller) workflowsclient.Client {
				return workflowsclient.NewMockClient(ctl)
			},
			errExpected: err,
		},
		"app repo error": {
			installRepo: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				mockRepo := repos.NewMockInstallRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), installID).Return(install, nil)
				return mockRepo
			},
			appRepo: func(ctl *gomock.Controller) *repos.MockAppRepo {
				mockRepo := repos.NewMockAppRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), install.AppID).Return(nil, err)
				return mockRepo
			},
			adminRepo: func(ctl *gomock.Controller) *repos.MockAdminRepo {
				mockRepo := repos.NewMockAdminRepo(ctl)
				return mockRepo
			},
			mockWfc: func(ctl *gomock.Controller) workflowsclient.Client {
				return workflowsclient.NewMockClient(ctl)
			},
			errExpected: err,
		},
		"admin repo error": {
			installRepo: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				mockRepo := repos.NewMockInstallRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), installID).Return(install, nil)
				return mockRepo
			},
			appRepo: func(ctl *gomock.Controller) *repos.MockAppRepo {
				mockRepo := repos.NewMockAppRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), install.AppID).Return(app, nil)
				return mockRepo
			},
			adminRepo: func(ctl *gomock.Controller) *repos.MockAdminRepo {
				mockRepo := repos.NewMockAdminRepo(ctl)
				mockRepo.EXPECT().GetLatestSandboxVersion(gomock.Any()).Return(nil, err)
				return mockRepo
			},
			mockWfc: func(ctl *gomock.Controller) workflowsclient.Client {
				return workflowsclient.NewMockClient(ctl)
			},
			errExpected: err,
		},
		"workflow error": {
			installRepo: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				mockRepo := repos.NewMockInstallRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), installID).Return(install, nil)
				return mockRepo
			},
			appRepo: func(ctl *gomock.Controller) *repos.MockAppRepo {
				mockRepo := repos.NewMockAppRepo(ctl)
				mockRepo.EXPECT().Get(gomock.Any(), install.AppID).Return(app, nil)
				return mockRepo
			},
			adminRepo: func(ctl *gomock.Controller) *repos.MockAdminRepo {
				mockRepo := repos.NewMockAdminRepo(ctl)
				mockRepo.EXPECT().GetLatestSandboxVersion(gomock.Any()).Return(sandboxVersion, nil)
				return mockRepo
			},
			mockWfc: func(ctl *gomock.Controller) workflowsclient.Client {
				mockWfc := workflowsclient.NewMockClient(ctl)
				mockWfc.EXPECT().TriggerInstallProvision(gomock.Any(), gomock.Any()).Return("", err)
				return mockWfc
			},
			errExpected: err,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			act := &activities{
				repo:      test.installRepo(mockCtl),
				appRepo:   test.appRepo(mockCtl),
				adminRepo: test.adminRepo(mockCtl),
				wfc:       test.mockWfc(mockCtl),
			}

			_, err := act.TriggerInstallJob(context.Background(), installID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
