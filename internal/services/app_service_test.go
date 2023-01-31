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
	"github.com/powertoolsdev/go-generics"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestAppService_GetApp(t *testing.T) {
	errGetApp := fmt.Errorf("error getting app")
	appID := uuid.New()
	app := generics.GetFakeObj[*models.App]()

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
			svc := &appService{
				log:  zaptest.NewLogger(t),
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

func TestAppService_UpsertApp(t *testing.T) {
	errUpsertApp := fmt.Errorf("error upserting app")
	app := generics.GetFakeObj[*models.App]()

	tests := map[string]struct {
		inputFn     func() models.AppInput
		repoFn      func(*gomock.Controller) *repos.MockAppRepo
		wkflowFn    func(*gomock.Controller) *workflows.MockAppWorkflowManager
		errExpected error
	}{
		"create a new app": {
			inputFn: func() models.AppInput {
				inp := generics.GetFakeObj[models.AppInput]()
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
		"upsert error": {
			inputFn: func() models.AppInput {
				inp := generics.GetFakeObj[models.AppInput]()
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
				inp := generics.GetFakeObj[models.AppInput]()
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
			svc := &appService{
				log:         zaptest.NewLogger(t),
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
			svc := &appService{
				log:  zaptest.NewLogger(t),
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
