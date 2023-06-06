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
	"github.com/powertoolsdev/mono/services/api/internal/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestComponentService_UpsertComponent(t *testing.T) {
	errUpsertComponent := fmt.Errorf("error upserting component")
	component := generics.GetFakeObj[*models.Component]()
	app := generics.GetFakeObj[*models.App]()
	app.ID = component.AppID

	tests := map[string]struct {
		inputFn     func() models.ComponentInput
		repoFn      func(*gomock.Controller) *repos.MockComponentRepo
		appRepoFn   func(*gomock.Controller) *repos.MockAppRepo
		errExpected error
	}{
		"create a new component": {
			inputFn: func() models.ComponentInput {
				inp := generics.GetFakeObj[models.ComponentInput]()
				inp.ID = nil
				inp.AppID = app.ID
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockComponentRepo {
				repo := repos.NewMockComponentRepo(ctl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(component, nil)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), component.AppID).Return(app, nil)
				return repo
			},
		},
		"app not found": {
			inputFn: func() models.ComponentInput {
				inp := generics.GetFakeObj[models.ComponentInput]()
				inp.ID = nil
				inp.AppID = app.ID
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockComponentRepo {
				repo := repos.NewMockComponentRepo(ctl)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), component.AppID).Return(nil, errUpsertComponent)
				return repo
			},
			errExpected: errUpsertComponent,
		},
		"upsert not found": {
			inputFn: func() models.ComponentInput {
				inp := generics.GetFakeObj[models.ComponentInput]()
				inp.ID = generics.ToPtr(component.ID)
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockComponentRepo {
				repo := repos.NewMockComponentRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), component.ID).Return(nil, errUpsertComponent)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				return repo
			},
			errExpected: errUpsertComponent,
		},
		"upsert without component config": {
			inputFn: func() models.ComponentInput {
				inp := generics.GetFakeObj[models.ComponentInput]()
				inp.ID = generics.ToPtr(component.ID)
				inp.Config = nil
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockComponentRepo {
				repo := repos.NewMockComponentRepo(ctl)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(component, nil)
				repo.EXPECT().Get(gomock.Any(), component.ID).Return(component, nil)
				return repo
			},
			appRepoFn: func(ctl *gomock.Controller) *repos.MockAppRepo {
				repo := repos.NewMockAppRepo(ctl)
				return repo
			},
			errExpected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			componentInput := test.inputFn()
			repo := test.repoFn(mockCtl)
			svc := &componentService{
				log:     zaptest.NewLogger(t),
				appRepo: test.appRepoFn(mockCtl),
				repo:    repo,
			}

			returnedComponent, err := svc.UpsertComponent(context.Background(), componentInput)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedComponent)
		})
	}
}

func TestComponentService_GetAppComponents(t *testing.T) {
	errGetAppComponents := fmt.Errorf("error getting app components")
	appID := domains.NewAppID()
	component := generics.GetFakeObj[*models.Component]()
	component.AppID = appID

	tests := map[string]struct {
		appID       string
		repoFn      func(*gomock.Controller) *repos.MockComponentRepo
		errExpected error
		assertFn    func(*testing.T, *models.Component)
	}{
		"happy path": {
			appID: appID,
			repoFn: func(ctl *gomock.Controller) *repos.MockComponentRepo {
				repo := repos.NewMockComponentRepo(ctl)
				repo.EXPECT().ListByApp(gomock.Any(), appID, &models.ConnectionOptions{}).Return([]*models.Component{component}, &utils.Page{}, nil)
				return repo
			},
		},
		"error": {
			appID: appID,
			repoFn: func(ctl *gomock.Controller) *repos.MockComponentRepo {
				repo := repos.NewMockComponentRepo(ctl)
				repo.EXPECT().ListByApp(gomock.Any(), appID, &models.ConnectionOptions{}).Return(nil, nil, errGetAppComponents)
				return repo
			},
			errExpected: errGetAppComponents,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &componentService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}

			components, _, err := svc.GetAppComponents(context.Background(), test.appID, &models.ConnectionOptions{})
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, component, components[0])
		})
	}
}

func TestComponentService_GetComponent(t *testing.T) {
	errGetComponent := fmt.Errorf("error getting component")
	componentID := domains.NewComponentID()
	component := generics.GetFakeObj[*models.Component]()

	tests := map[string]struct {
		componentID string
		repoFn      func(*gomock.Controller) *repos.MockComponentRepo
		errExpected error
		assertFn    func(*testing.T, *models.Component)
	}{
		"happy path": {
			componentID: componentID,
			repoFn: func(ctl *gomock.Controller) *repos.MockComponentRepo {
				repo := repos.NewMockComponentRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), componentID).Return(component, nil)
				return repo
			},
		},
		"error": {
			componentID: componentID,
			repoFn: func(ctl *gomock.Controller) *repos.MockComponentRepo {
				repo := repos.NewMockComponentRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), componentID).Return(nil, errGetComponent)
				return repo
			},
			errExpected: errGetComponent,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &componentService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}
			returnedComponent, err := svc.GetComponent(context.Background(), test.componentID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedComponent)
		})
	}
}

func TestComponentService_DeleteComponent(t *testing.T) {
	errDeleteComponent := fmt.Errorf("error deleting component")
	componentID := domains.NewComponentID()

	tests := map[string]struct {
		componentID string
		repoFn      func(*gomock.Controller) *repos.MockComponentRepo
		errExpected error
	}{
		"happy path": {
			componentID: componentID,
			repoFn: func(ctl *gomock.Controller) *repos.MockComponentRepo {
				repo := repos.NewMockComponentRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), componentID).Return(true, nil)
				return repo
			},
		},
		"delete error": {
			componentID: componentID,
			repoFn: func(ctl *gomock.Controller) *repos.MockComponentRepo {
				repo := repos.NewMockComponentRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), componentID).Return(false, errDeleteComponent)
				return repo
			},
			errExpected: errDeleteComponent,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &componentService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}

			returnedComponent, err := svc.DeleteComponent(context.Background(), test.componentID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedComponent)
		})
	}
}
