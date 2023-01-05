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

func TestAppService_GetApp(t *testing.T) {
	errGetApp := fmt.Errorf("error getting app")
	appID := uuid.New()
	app := getFakeObj[*models.App]()

	tests := map[string]struct {
		appID       string
		repoFn      func(*gomock.Controller) *repos.MockAppRepo
		errExpected error
		assertFn    func(*testing.T, *models.App)
	}{
		"happy path": {
			appID: appID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), appID).Return(app, nil)
				return repo
			},
		},
		"invalid-id": {
			appID: "foo",
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				return repo
			},
			errExpected: fmt.Errorf("is not a valid uuid"),
		},
		"error": {
			appID: appID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), appID).Return(nil, errGetApp)
				return repo
			},
			errExpected: errGetApp,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &AppService{
				repo: repo,
			}
			returnedApp, err := svc.GetApp(context.Background(), test.appID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedApp)
		})
	}
}

func TestAppService_GetAppBySlug(t *testing.T) {
	errGetApp := fmt.Errorf("error getting app")
	slug := uuid.NewString()
	app := getFakeObj[*models.App]()

	tests := map[string]struct {
		slug        string
		repoFn      func(*gomock.Controller) *repos.MockAppRepo
		errExpected error
		assertFn    func(*testing.T, *models.App)
	}{
		"happy path": {
			slug: slug,
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().GetBySlug(gomock.Any(), slug).Return(app, nil)
				return repo
			},
		},
		"error": {
			slug: "invalid-slug",
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().GetBySlug(gomock.Any(), "invalid-slug").Return(nil, errGetApp)
				return repo
			},
			errExpected: errGetApp,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &AppService{
				repo: repo,
			}
			returnedApp, err := svc.GetAppBySlug(context.Background(), test.slug)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedApp)
		})
	}
}

func TestAppService_UpsertApp(t *testing.T) {
	errUpsertApp := fmt.Errorf("error upserting app")
	app := getFakeObj[*models.App]()

	tests := map[string]struct {
		inputFn     func() models.AppInput
		repoFn      func(*gomock.Controller) *repos.MockAppRepo
		wkflowFn    func(*gomock.Controller) *workflows.MockAppWorkflowManager
		errExpected error
	}{
		"create a new app": {
			inputFn: func() models.AppInput {
				inp := getFakeObj[models.AppInput]()
				inp.ID = nil
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(app, nil)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockAppWorkflowManager {
				mgr := workflows.NewMockAppWorkflowManager(ctl)
				mgr.EXPECT().Provision(gomock.Any(), app).Return(nil)
				return mgr
			},
		},
		"invalid id": {
			inputFn: func() models.AppInput {
				inp := getFakeObj[models.AppInput]()
				inp.ID = toPtr("abc")
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				return repos.NewMockAppRepo(ctl)
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockAppWorkflowManager {
				return workflows.NewMockAppWorkflowManager(ctl)
			},
			errExpected: InvalidIDErr{},
		},
		"upsert error": {
			inputFn: func() models.AppInput {
				inp := getFakeObj[models.AppInput]()
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil, errUpsertApp)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockAppWorkflowManager {
				return workflows.NewMockAppWorkflowManager(ctl)
			},
			errExpected: errUpsertApp,
		},
		"error provisioning": {
			inputFn: func() models.AppInput {
				inp := getFakeObj[models.AppInput]()
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(app, nil)
				return repo
			},
			wkflowFn: func(ctl *gomock.Controller) *workflows.MockAppWorkflowManager {
				mgr := workflows.NewMockAppWorkflowManager(ctl)
				mgr.EXPECT().Provision(gomock.Any(), gomock.Any()).Return(errUpsertApp)
				return mgr
			},
			errExpected: errUpsertApp,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			appInput := test.inputFn()
			repo := test.repoFn(mockCtl)
			mgr := test.wkflowFn(mockCtl)
			svc := &AppService{
				repo:        repo,
				workflowMgr: mgr,
			}

			returnedApp, err := svc.UpsertApp(context.Background(), appInput)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedApp)
		})
	}
}

func TestAppService_DeleteApp(t *testing.T) {
	errDeleteApp := fmt.Errorf("error deleting app")
	appID := uuid.New()

	tests := map[string]struct {
		appID       string
		repoFn      func(*gomock.Controller) *repos.MockAppRepo
		errExpected error
	}{
		"happy path": {
			appID: appID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), appID).Return(true, nil)
				return repo
			},
		},
		"invalid id": {
			appID: "invalid-id",
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				return repos.NewMockAppRepo(ctl)
			},
			errExpected: InvalidIDErr{},
		},
		"delete error": {
			appID: appID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), appID).Return(false, errDeleteApp)
				return repo
			},
			errExpected: errDeleteApp,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &AppService{
				repo: repo,
			}

			returnedApp, err := svc.DeleteApp(context.Background(), test.appID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedApp)
		})
	}
}

func TestAppService_GetAllApps(t *testing.T) {
	errGetAllApps := fmt.Errorf("error getting all apps")
	apps := []*models.App{getFakeObj[*models.App]()}
	opts := &models.ConnectionOptions{}

	tests := map[string]struct {
		repoFn      func(*gomock.Controller) *repos.MockAppRepo
		errExpected error
	}{
		"happy path": {
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().GetPageAll(gomock.Any(), opts).Return(apps, nil, nil)
				return repo
			},
		},
		"repo error": {
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().GetPageAll(gomock.Any(), opts).Return(nil, nil, errGetAllApps)
				return repo
			},
			errExpected: errGetAllApps,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &AppService{
				repo: repo,
			}

			returnedApps, _, err := svc.GetAllApps(context.Background(), opts)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, apps, returnedApps)
		})
	}
}
