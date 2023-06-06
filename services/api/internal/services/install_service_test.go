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

func TestInstallService_UpsertInstall(t *testing.T) {
	errUpsertInstall := fmt.Errorf("error upserting install")
	install := generics.GetFakeObj[*models.Install]()

	tests := map[string]struct {
		inputFn     func() models.InstallInput
		repoFn      func(*gomock.Controller) *repos.MockInstallRepo
		errExpected error
	}{
		"create a new install": {
			inputFn: func() models.InstallInput {
				inp := generics.GetFakeObj[models.InstallInput]()
				inp.ID = nil
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(install, nil)
				return repo
			},
		},
		"update an install": {
			inputFn: func() models.InstallInput {
				inp := generics.GetFakeObj[models.InstallInput]()
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(install, nil)
				repo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(install, nil)
				return repo
			},
		},
		"create error": {
			inputFn: func() models.InstallInput {
				inp := generics.GetFakeObj[models.InstallInput]()
				inp.ID = nil
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errUpsertInstall)
				return repo
			},
			errExpected: errUpsertInstall,
		},
		"update error": {
			inputFn: func() models.InstallInput {
				inp := generics.GetFakeObj[models.InstallInput]()
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errUpsertInstall)
				repo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(install, nil)
				return repo
			},
			errExpected: errUpsertInstall,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			appInput := test.inputFn()
			svc := &installService{
				log:  zaptest.NewLogger(t),
				repo: test.repoFn(mockCtl),
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
	appID := domains.NewAppID()
	install := generics.GetFakeObj[*models.Install]()

	tests := map[string]struct {
		appID       string
		repoFn      func(*gomock.Controller) *repos.MockInstallRepo
		errExpected error
		assertFn    func(*testing.T, *models.Install)
	}{
		"happy path": {
			appID: appID,
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().ListByApp(gomock.Any(), appID, &models.ConnectionOptions{}).Return([]*models.Install{install}, nil, nil)
				return repo
			},
		},
		"error": {
			appID: appID,
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
			svc := &installService{
				log:  zaptest.NewLogger(t),
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
	installID := domains.NewInstallID()
	app := generics.GetFakeObj[*models.Install]()

	tests := map[string]struct {
		installID   string
		repoFn      func(*gomock.Controller) *repos.MockInstallRepo
		errExpected error
		assertFn    func(*testing.T, *models.Install)
	}{
		"happy path": {
			installID: installID,
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), installID).Return(app, nil)
				return repo
			},
		},
		"error": {
			installID: installID,
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
			svc := &installService{
				log:  zaptest.NewLogger(t),
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
	installID := domains.NewInstallID()

	tests := map[string]struct {
		installID   string
		repoFn      func(*gomock.Controller) *repos.MockInstallRepo
		errExpected error
	}{
		"happy path": {
			installID: installID,
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), installID).Return(true, nil)
				return repo
			},
		},
		"error deleting install": {
			installID: installID,
			repoFn: func(ctl *gomock.Controller) *repos.MockInstallRepo {
				repo := repos.NewMockInstallRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), installID).Return(false, errDeleteInstall)
				return repo
			},
			errExpected: errDeleteInstall,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			svc := &installService{
				log:  zaptest.NewLogger(t),
				repo: test.repoFn(mockCtl),
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
