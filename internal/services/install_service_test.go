package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/workflows"
	"github.com/stretchr/testify/assert"
)

func TestInstallService_UpsertInstall(t *testing.T) {
	errUpsertInstall := fmt.Errorf("error upserting install")
	install := getFakeObj[*models.Install]()
	app := getFakeObj[*models.App]()
	fmt.Println(errUpsertInstall)

	tests := map[string]struct {
		inputFn     func() models.InstallInput
		repoFn      func(*gomock.Controller) *repos.MockInstallRepo
		appRepoFn   func(*gomock.Controller) *repos.MockAppRepo
		wkflowFn    func(*gomock.Controller) *workflows.MockInstallWorkflowManager
		errExpected error
	}{
		"create a new app": {
			inputFn: func() models.InstallInput {
				inp := getFakeObj[models.InstallInput]()
				inp.ID = nil
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(install, nil)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), install.AppID).Return(app, nil)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				mgr := workflows.NewMockInstallWorkflowManager(ctl)
				mgr.EXPECT().Provision(gomock.Any(), install, app.OrgID.String()).Return(nil)
				return mgr
			},
		},
		"update an app": {
			inputFn: func() models.InstallInput {
				inp := getFakeObj[models.InstallInput]()
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(install, nil)
				repo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(install, nil)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), install.AppID).Return(app, nil)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				mgr := workflows.NewMockInstallWorkflowManager(ctl)
				mgr.EXPECT().Provision(gomock.Any(), install, app.OrgID.String()).Return(nil)
				return mgr
			},
		},
		"update with invalid id": {
			inputFn: func() models.InstallInput {
				inp := getFakeObj[models.InstallInput]()
				inp.ID = toPtr("abc")
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				return repos.NewMockInstallRepo(ctl)
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				return workflows.NewMockInstallWorkflowManager(ctl)
			},
			errExpected: InvalidIDErr{},
		},
		"create error": {
			inputFn: func() models.InstallInput {
				inp := getFakeObj[models.InstallInput]()
				inp.ID = nil
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errUpsertInstall)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				return workflows.NewMockInstallWorkflowManager(ctl)
			},
			errExpected: errUpsertInstall,
		},
		"update error": {
			inputFn: func() models.InstallInput {
				inp := getFakeObj[models.InstallInput]()
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errUpsertInstall)
				repo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(install, nil)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				return workflows.NewMockInstallWorkflowManager(ctl)
			},
			errExpected: errUpsertInstall,
		},
		"error provisioning on create": {
			inputFn: func() models.InstallInput {
				inp := getFakeObj[models.InstallInput]()
				inp.ID = nil
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(install, nil)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), install.AppID).Return(app, nil)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				mgr := workflows.NewMockInstallWorkflowManager(ctl)
				mgr.EXPECT().Provision(gomock.Any(), install, app.OrgID.String()).Return(errUpsertInstall)
				return mgr
			},
			errExpected: errUpsertInstall,
		},
		"error provisioning on update": {
			inputFn: func() models.InstallInput {
				inp := getFakeObj[models.InstallInput]()
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(install, nil)
				repo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(install, nil)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), install.AppID).Return(app, nil)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				mgr := workflows.NewMockInstallWorkflowManager(ctl)
				mgr.EXPECT().Provision(gomock.Any(), install, app.OrgID.String()).Return(errUpsertInstall)
				return mgr
			},
			errExpected: errUpsertInstall,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			appInput := test.inputFn()
			svc := &InstallService{
				repo:      test.repoFn(mockCtl),
				wkflowMgr: test.wkflowFn(mockCtl),
				appRepo:   test.appRepoFn(mockCtl),
			}

			returnedInstall, err := svc.UpsertInstall(context.Background(), appInput)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedInstall)
		})
	}
}

func TestInstallService_GetAppInstalls(t *testing.T) {
	errGetAppInstalls := fmt.Errorf("error getting app installs")
	appID := uuid.New()
	install := getFakeObj[*models.Install]()

	tests := map[string]struct {
		appID       string
		repoFn      func(*gomock.Controller) *repos.MockInstallRepo
		errExpected error
		assertFn    func(*testing.T, *models.Install)
	}{
		"happy path": {
			appID: appID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().ListByApp(gomock.Any(), appID, &models.ConnectionOptions{}).Return([]*models.Install{install}, nil, nil)
				return repo
			},
		},
		"invalid-id": {
			appID: "foo",
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				return repo
			},
			errExpected: fmt.Errorf("is not a valid uuid"),
		},
		"error": {
			appID: appID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().ListByApp(gomock.Any(), appID, &models.ConnectionOptions{}).Return(nil, nil, errGetAppInstalls)
				return repo
			},
			errExpected: errGetAppInstalls,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &InstallService{
				repo: repo,
			}
			installs, _, err := svc.GetAppInstalls(context.Background(), test.appID, &models.ConnectionOptions{})
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, install, installs[0])
		})
	}
}

func TestInstallService_GetInstall(t *testing.T) {
	errGetInstall := fmt.Errorf("error getting install")
	installID := uuid.New()
	app := getFakeObj[*models.Install]()

	tests := map[string]struct {
		installID   string
		repoFn      func(*gomock.Controller) *repos.MockInstallRepo
		errExpected error
		assertFn    func(*testing.T, *models.Install)
	}{
		"happy path": {
			installID: installID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), installID).Return(app, nil)
				return repo
			},
		},
		"invalid-id": {
			installID: "foo",
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				return repo
			},
			errExpected: fmt.Errorf("is not a valid uuid"),
		},
		"error": {
			installID: installID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), installID).Return(nil, errGetInstall)
				return repo
			},
			errExpected: errGetInstall,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &InstallService{
				repo: repo,
			}
			returnedInstall, err := svc.GetInstall(context.Background(), test.installID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedInstall)
		})
	}
}

func TestInstallService_DeleteInstall(t *testing.T) {
	errDeleteInstall := fmt.Errorf("error deleting install")
	installID := uuid.New()
	install := getFakeObj[*models.Install]()
	app := getFakeObj[*models.App]()

	tests := map[string]struct {
		installID   string
		repoFn      func(*gomock.Controller) *repos.MockInstallRepo
		appRepoFn   func(*gomock.Controller) *repos.MockAppRepo
		wkflowFn    func(*gomock.Controller) *workflows.MockInstallWorkflowManager
		errExpected error
	}{
		"happy path": {
			installID: installID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), installID).Return(true, nil)
				repo.EXPECT().Get(gomock.Any(), installID).Return(install, nil)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), install.AppID).Return(app, nil)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				mgr := workflows.NewMockInstallWorkflowManager(ctl)
				mgr.EXPECT().Deprovision(gomock.Any(), install, app.OrgID.String()).Return(nil)
				return mgr
			},
		},
		"invalid id": {
			installID: "invalid-id",
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				return repos.NewMockInstallRepo(ctl)
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				mgr := workflows.NewMockInstallWorkflowManager(ctl)
				return mgr
			},
			errExpected: InvalidIDErr{},
		},
		"error fetching app": {
			installID: installID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), installID).Return(true, nil)
				repo.EXPECT().Get(gomock.Any(), installID).Return(install, nil)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), install.AppID).Return(nil, errDeleteInstall)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				mgr := workflows.NewMockInstallWorkflowManager(ctl)
				return mgr
			},
			errExpected: errDeleteInstall,
		},
		"error deleting install": {
			installID: installID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), installID).Return(false, errDeleteInstall)
				repo.EXPECT().Get(gomock.Any(), installID).Return(install, nil)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				mgr := workflows.NewMockInstallWorkflowManager(ctl)
				return mgr
			},
			errExpected: errDeleteInstall,
		},
		"error deprovisioning": {
			installID: installID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), installID).Return(true, nil)
				repo.EXPECT().Get(gomock.Any(), installID).Return(install, nil)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), install.AppID).Return(app, nil)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockInstallWorkflowManager {
				mgr := workflows.NewMockInstallWorkflowManager(ctl)
				mgr.EXPECT().Deprovision(gomock.Any(), install, app.OrgID.String()).Return(errDeleteInstall)
				return mgr
			},
			errExpected: errDeleteInstall,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			svc := &InstallService{
				repo:      test.repoFn(mockCtl),
				appRepo:   test.appRepoFn(mockCtl),
				wkflowMgr: test.wkflowFn(mockCtl),
			}

			deleted, err := svc.DeleteInstall(context.Background(), test.installID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.True(t, deleted)
		})
	}
}
