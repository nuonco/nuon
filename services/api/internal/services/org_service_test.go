package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestOrgService_DeleteOrg(t *testing.T) {
	errDeleteOrg := fmt.Errorf("error deleting org")
	orgID := domains.NewOrgID()

	tests := map[string]struct {
		orgID       string
		repoFn      func(*gomock.Controller) *repos.MockOrgRepo
		errExpected error
	}{
		"happy path": {
			orgID: orgID,
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), orgID).Return(true, nil)
				return repo
			},
		},
		"delete error": {
			orgID: orgID,
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), orgID).Return(false, errDeleteOrg)
				return repo
			},
			errExpected: errDeleteOrg,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &orgService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}

			deleted, err := svc.DeleteOrg(context.Background(), test.orgID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.True(t, deleted)
		})
	}
}

func TestOrgService_GetOrg(t *testing.T) {
	errGetOrg := fmt.Errorf("error getting org")
	orgID := domains.NewOrgID()
	org := generics.GetFakeObj[*models.Org]()

	tests := map[string]struct {
		orgID       string
		repoFn      func(*gomock.Controller) *repos.MockOrgRepo
		errExpected error
		assertFn    func(*testing.T, *models.Org)
	}{
		"happy path": {
			orgID: orgID,
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), orgID).Return(org, nil)
				return repo
			},
		},
		"error": {
			orgID: orgID,
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), orgID).Return(nil, errGetOrg)
				return repo
			},
			errExpected: errGetOrg,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &orgService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}
			returnedOrg, err := svc.GetOrg(context.Background(), test.orgID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, org, returnedOrg)
		})
	}
}

func TestOrgService_UpsertOrg(t *testing.T) {
	errUpsertOrg := fmt.Errorf("error upserting org")
	org := generics.GetFakeObj[*models.Org]()
	org.IsNew = false
	org.GithubInstallID = "1234567890"
	userID := uuid.NewString()

	tests := map[string]struct {
		inputFn     func() models.OrgInput
		repoFn      func(*gomock.Controller) *repos.MockOrgRepo
		userRepoFn  func(*gomock.Controller) *repos.MockUserRepo
		errExpected error
	}{
		"create happy path": {
			inputFn: func() models.OrgInput {
				inp := generics.GetFakeObj[models.OrgInput]()
				inp.ID = nil
				inp.OwnerID = userID
				inp.GithubInstallID = &org.GithubInstallID
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				returnedOrg := generics.GetFakeObj[*models.Org]()
				returnedOrg.IsNew = true
				returnedOrg.ID = org.ID
				returnedOrg.GithubInstallID = org.GithubInstallID
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(returnedOrg, nil)
				return repo
			},
			userRepoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().UpsertUserOrg(gomock.Any(), userID, org.ID).Return(&models.UserOrg{}, nil)
				return repo
			},
		},
		"update happy path": {
			inputFn: func() models.OrgInput {
				inp := generics.GetFakeObj[models.OrgInput]()
				inp.ID = generics.ToPtr(org.ID)
				inp.GithubInstallID = &org.GithubInstallID
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				returnedOrg := generics.GetFakeObj[*models.Org]()
				returnedOrg.IsNew = false
				repo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(returnedOrg, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(returnedOrg, nil)
				return repo
			},
			userRepoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				return repo
			},
		},
		"repo error": {
			inputFn: func() models.OrgInput {
				inp := generics.GetFakeObj[models.OrgInput]()
				inp.ID = nil
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errUpsertOrg)
				return repo
			},
			userRepoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				return repo
			},
			errExpected: errUpsertOrg,
		},
		//"upsert user org error": {
		//ctxFn: func() context.Context {
		//ctx := context.Background()
		//ctx = context.WithValue(ctx, middleware.UserContext{}, user)
		//return ctx
		//},
		//inputFn: func() models.OrgInput {
		//inp := generics.GetFakeObj[models.OrgInput]()
		//inp.ID = nil
		//return inp
		//},
		//repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
		//repo := repos.NewMockOrgRepo(ctl)
		//repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(org, nil)
		//return repo
		//},
		//userRepoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
		//repo := repos.NewMockUserRepo(ctl)
		//repo.EXPECT().UpsertUserOrg(gomock.Any(), user.ID, org.ID).Return(nil, errUpsertOrg)
		//return repo
		//},
		//wkflowFn: func(ctl *gomock.Controller) *workflows.MockOrgWorkflowManager {
		//mgr := workflows.NewMockOrgWorkflowManager(ctl)
		////mgr.EXPECT().Provision(gomock.Any(), org.ID).Return(nil)
		//return mgr
		//},
		//errExpected: errUpsertOrg,
		//},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			orgInput := test.inputFn()
			svc := &orgService{
				log:            zaptest.NewLogger(t),
				repo:           test.repoFn(mockCtl),
				userOrgUpdater: test.userRepoFn(mockCtl),
			}

			ctx := context.Background()
			returnedOrg, err := svc.UpsertOrg(ctx, orgInput)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedOrg)
		})
	}
}

func TestOrgService_UserOrgs(t *testing.T) {
	errGetUserOrgs := fmt.Errorf("error getting user orgs")
	userID := "123"
	org := generics.GetFakeObj[*models.Org]()

	tests := map[string]struct {
		userID      string
		repoFn      func(*gomock.Controller) *repos.MockOrgRepo
		errExpected error
		assertFn    func(*testing.T, *models.Org)
	}{
		"happy path": {
			userID: userID,
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				repo.EXPECT().GetPageByUser(gomock.Any(), userID, gomock.Any()).Return([]*models.Org{org}, nil, nil)
				return repo
			},
		},
		"error": {
			userID: userID,
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				repo.EXPECT().GetPageByUser(gomock.Any(), userID, gomock.Any()).Return(nil, nil, errGetUserOrgs)
				return repo
			},
			errExpected: errGetUserOrgs,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &orgService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}
			orgs, _, err := svc.UserOrgs(context.Background(), test.userID, &models.ConnectionOptions{})
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, org, orgs[0])
		})
	}
}
