package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestAppService_GetApp(t *testing.T) {
	errGetApp := fmt.Errorf("error getting app")
	appID := domains.NewAppID()
	app := generics.GetFakeObj[*models.App]()

	tests := map[string]struct {
		appID       string
		repoFn      func(*gomock.Controller) *repos.MockAppRepo
		errExpected error
		assertFn    func(*testing.T, *models.App)
	}{
		"happy path": {
			appID: appID,
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), appID).Return(app, nil)
				return repo
			},
		},
		"error": {
			appID: appID,
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

func TestAppService_GetOrgApps(t *testing.T) {
	errGetOrgApps := fmt.Errorf("error getting apps")
	orgID := domains.NewOrgID()
	app := generics.GetFakeObj[*models.App]()

	tests := map[string]struct {
		orgID       string
		repoFn      func(*gomock.Controller) *repos.MockAppRepo
		errExpected error
		assertFn    func(*testing.T, *models.App)
	}{
		"happy path": {
			orgID: orgID,
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().GetPageByOrg(gomock.Any(), orgID, gomock.Any()).Return([]*models.App{app}, nil, nil)
				return repo
			},
		},
		"error": {
			orgID: orgID,
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().GetPageByOrg(gomock.Any(), orgID, gomock.Any()).Return(nil, nil, errGetOrgApps)
				return repo
			},
			errExpected: errGetOrgApps,
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
			returnedApp, _, err := svc.GetOrgApps(context.Background(), test.orgID, &models.ConnectionOptions{})
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
	org := generics.GetFakeObj[*models.Org]()
	org.ID = app.OrgID

	tests := map[string]struct {
		inputFn     func() models.AppInput
		orgRepoFn   func(*gomock.Controller) *repos.MockOrgRepo
		repoFn      func(*gomock.Controller) *repos.MockAppRepo
		errExpected error
	}{
		"create a new app": {
			inputFn: func() models.AppInput {
				inp := generics.GetFakeObj[models.AppInput]()
				inp.ID = nil
				inp.OrgID = org.ID
				return inp
			},
			orgRepoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), app.OrgID).Return(org, nil)
				return repo
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(app, nil)
				return repo
			},
		},
		"upsert happy path": {
			inputFn: func() models.AppInput {
				inp := generics.GetFakeObj[models.AppInput]()
				inp.ID = generics.ToPtr(app.ID)
				inp.OrgID = org.ID
				return inp
			},
			orgRepoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), app.OrgID).Return(org, nil)
				return repo
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(app, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(app, nil)
				return repo
			},
		},
		"org not found": {
			inputFn: func() models.AppInput {
				inp := generics.GetFakeObj[models.AppInput]()
				inp.ID = generics.ToPtr(app.ID)
				inp.OrgID = org.ID
				return inp
			},
			orgRepoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), app.OrgID).Return(nil, errUpsertApp)
				return repo
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				return repo
			},
			errExpected: errUpsertApp,
		},
		"upsert not found": {
			inputFn: func() models.AppInput {
				inp := generics.GetFakeObj[models.AppInput]()
				inp.ID = generics.ToPtr(app.ID)
				inp.OrgID = org.ID
				return inp
			},
			orgRepoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), app.OrgID).Return(org, nil)
				return repo
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), app.ID).Return(nil, errUpsertApp)
				return repo
			},
			errExpected: errUpsertApp,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			appInput := test.inputFn()
			repo := test.repoFn(mockCtl)
			svc := &appService{
				log:     zaptest.NewLogger(t),
				orgRepo: test.orgRepoFn(mockCtl),
				repo:    repo,
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
	appID := domains.NewAppID()

	tests := map[string]struct {
		appID       string
		repoFn      func(*gomock.Controller) *repos.MockAppRepo
		errExpected error
	}{
		"happy path": {
			appID: appID,
			repoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), appID).Return(true, nil)
				return repo
			},
		},
		"delete error": {
			appID: appID,
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
