package deleteinstall

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

func Test_ActivityTriggerInstallDeprovJob(t *testing.T) {
	err := errors.New("error")
	app := generics.GetFakeObj[*models.App]()

	installID, _ := shortid.NewNanoID("inl")
	install := generics.GetFakeObj[*models.Install]()
	sandboxVersion := generics.GetFakeObj[*models.SandboxVersion]()

	tests := map[string]struct {
		installRepo func(*gomock.Controller) *repos.MockInstallRepo
		appRepo     func(*gomock.Controller) *repos.MockAppRepo
		adminRepo   func(*gomock.Controller) *repos.MockAdminRepo
		mockMgr     func(*gomock.Controller) *workflows.MockInstallWorkflowManager
		errExpected error
	}{
		"happy path": {
			installRepo: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				mockRepo := repos.NewMockInstallRepo(ctl)
				mockRepo.EXPECT().GetDeleted(gomock.Any(), installID).Return(install, nil)
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
			mockMgr: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				wkflowmgr := workflows.NewMockInstallWorkflowManager(ctl)
				wkflowmgr.EXPECT().Deprovision(gomock.Any(), install, app.OrgID, sandboxVersion).Return("12354", nil)
				return wkflowmgr
			},
		},
		"repo error": {
			installRepo: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				mockRepo := repos.NewMockInstallRepo(ctl)
				mockRepo.EXPECT().GetDeleted(gomock.Any(), installID).Return(nil, err)
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
			mockMgr: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				wkflowmgr := workflows.NewMockInstallWorkflowManager(ctl)
				return wkflowmgr
			},
			errExpected: err,
		},
		"app repo error": {
			installRepo: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				mockRepo := repos.NewMockInstallRepo(ctl)
				mockRepo.EXPECT().GetDeleted(gomock.Any(), installID).Return(install, nil)
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
			mockMgr: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				wkflowmgr := workflows.NewMockInstallWorkflowManager(ctl)
				return wkflowmgr
			},
			errExpected: err,
		},
		"admin repo error": {
			installRepo: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				mockRepo := repos.NewMockInstallRepo(ctl)
				mockRepo.EXPECT().GetDeleted(gomock.Any(), installID).Return(install, nil)
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
			mockMgr: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				wkflowmgr := workflows.NewMockInstallWorkflowManager(ctl)
				return wkflowmgr
			},
			errExpected: err,
		},
		"workflow error": {
			installRepo: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				mockRepo := repos.NewMockInstallRepo(ctl)
				mockRepo.EXPECT().GetDeleted(gomock.Any(), installID).Return(install, nil)
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
			mockMgr: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				wkflowmgr := workflows.NewMockInstallWorkflowManager(ctl)
				wkflowmgr.EXPECT().Deprovision(gomock.Any(), install, app.OrgID, sandboxVersion).Return("", err)
				return wkflowmgr
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
				mgr:       test.mockMgr(mockCtl),
			}

			_, err := act.TriggerInstallDeprovJob(context.Background(), installID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
