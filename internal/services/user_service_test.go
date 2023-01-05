package services

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/utils"
	"github.com/powertoolsdev/go-generics"
	"github.com/stretchr/testify/assert"
)

func TestUserService_DeleteUser(t *testing.T) {
	errDeleteUser := fmt.Errorf("error deleting user")
	userID := uuid.New()

	tests := map[string]struct {
		userID      string
		repoFn      func(*gomock.Controller) *repos.MockUserRepo
		errExpected error
	}{
		"happy path": {
			userID: userID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), userID).Return(true, nil)
				return repo
			},
		},
		"invalid id": {
			userID: "invalid-id",
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				return repos.NewMockUserRepo(ctl)
			},
			errExpected: InvalidIDErr{},
		},
		"delete error": {
			userID: userID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().Delete(gomock.Any(), userID).Return(false, errDeleteUser)
				return repo
			},
			errExpected: errDeleteUser,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &UserService{
				repo: repo,
			}

			deleted, err := svc.DeleteUser(context.Background(), test.userID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				assert.False(t, deleted)
				return
			}
			assert.Nil(t, err)
			assert.True(t, deleted)
		})
	}
}

func TestUserService_UpsertUser(t *testing.T) {
	type fields struct {
		repo repos.UserRepo
	}
	type args struct {
		ctx   context.Context
		input models.UserInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserService{
				repo: tt.fields.repo,
			}
			got, err := u.UpsertUser(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserService.UpsertUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserService.UpsertUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserService_UpsertUserOrg(t *testing.T) {
	errUpsertUser := fmt.Errorf("error upserting user")
	user := generics.GetFakeObj[*models.User]()

	tests := map[string]struct {
		inputFn     func() models.UserInput
		repoFn      func(*gomock.Controller) *repos.MockUserRepo
		errExpected error
	}{
		"create a new app": {
			inputFn: func() models.UserInput {
				inp := generics.GetFakeObj[models.UserInput]()
				inp.ID = nil
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(user, nil)
				return repo
			},
		},
		"invalid id": {
			inputFn: func() models.UserInput {
				inp := generics.GetFakeObj[models.UserInput]()
				inp.ID = generics.ToPtr("abc")
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				return repos.NewMockUserRepo(ctl)
			},
			errExpected: InvalidIDErr{},
		},
		"upsert error": {
			inputFn: func() models.UserInput {
				inp := generics.GetFakeObj[models.UserInput]()
				return inp
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil, errUpsertUser)
				return repo
			},
			errExpected: errUpsertUser,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			userInput := test.inputFn()
			svc := &UserService{
				repo: test.repoFn(mockCtl),
			}

			returnedUser, err := svc.UpsertUser(context.Background(), userInput)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedUser)
		})
	}
}

func TestUserService_GetOrgUsers(t *testing.T) {
	errGetOrgUsers := fmt.Errorf("error getting org users")
	orgID := uuid.New()
	user := generics.GetFakeObj[*models.User]()

	tests := map[string]struct {
		orgID       string
		repoFn      func(*gomock.Controller) *repos.MockUserRepo
		errExpected error
		assertFn    func(*testing.T, *models.Org)
	}{
		"happy path": {
			orgID: orgID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().GetPageByOrg(gomock.Any(), orgID, &models.ConnectionOptions{}).Return([]*models.User{user}, nil, nil)
				return repo
			},
		},
		"invalid-id": {
			orgID: "foo",
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				return repo
			},
			errExpected: InvalidIDErr{},
		},
		"error": {
			orgID: orgID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().GetPageByOrg(gomock.Any(), orgID, &models.ConnectionOptions{}).Return(nil, nil, errGetOrgUsers)
				return repo
			},
			errExpected: errGetOrgUsers,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &UserService{
				repo: repo,
			}
			users, _, err := svc.GetOrgUsers(context.Background(), test.orgID, &models.ConnectionOptions{})
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, user, users[0])
		})
	}
}

func TestUserService_GetUserByEmail(t *testing.T) {
	errGetUser := fmt.Errorf("error getting user")
	email := "test@nuon.co"
	user := generics.GetFakeObj[*models.User]()

	tests := map[string]struct {
		email       string
		repoFn      func(*gomock.Controller) *repos.MockUserRepo
		errExpected error
		assertFn    func(*testing.T, *models.User)
	}{
		"happy path": {
			email: email,
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().GetByEmail(gomock.Any(), email).Return(user, nil)
				return repo
			},
		},
		"error": {
			email: email,
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().GetByEmail(gomock.Any(), email).Return(nil, errGetUser)
				return repo
			},
			errExpected: errGetUser,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &UserService{
				repo: repo,
			}
			returnedUser, err := svc.GetUserByEmail(context.Background(), test.email)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, user, returnedUser)
		})
	}
}

func TestUserService_GetUserByExternalID(t *testing.T) {
	errGetUser := fmt.Errorf("error getting user")
	externalID := uuid.NewString()
	user := generics.GetFakeObj[*models.User]()

	tests := map[string]struct {
		externalID  string
		repoFn      func(*gomock.Controller) *repos.MockUserRepo
		errExpected error
		assertFn    func(*testing.T, *models.User)
	}{
		"happy path": {
			externalID: externalID,
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().GetByExternalID(gomock.Any(), externalID).Return(user, nil)
				return repo
			},
		},
		"error": {
			externalID: externalID,
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().GetByExternalID(gomock.Any(), externalID).Return(nil, errGetUser)
				return repo
			},
			errExpected: errGetUser,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &UserService{
				repo: repo,
			}
			returnedUser, err := svc.GetUserByExternalID(context.Background(), test.externalID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, user, returnedUser)
		})
	}
}

func TestUserService_GetAllUsers(t *testing.T) {
	errGetAllUsers := fmt.Errorf("error getting all users")
	userID := uuid.New()
	user := generics.GetFakeObj[*models.User]()

	tests := map[string]struct {
		userID      string
		repoFn      func(*gomock.Controller) *repos.MockUserRepo
		errExpected error
		assertFn    func(*testing.T, *models.User)
	}{
		"happy path": {
			userID: userID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().GetPageAll(gomock.Any(), &models.ConnectionOptions{}).Return([]*models.User{user}, &utils.Page{}, nil)
				return repo
			},
		},
		"error": {
			userID: userID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().GetPageAll(gomock.Any(), &models.ConnectionOptions{}).Return(nil, nil, errGetAllUsers)
				return repo
			},
			errExpected: errGetAllUsers,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &UserService{
				repo: repo,
			}
			users, _, err := svc.GetAllUsers(context.Background(), &models.ConnectionOptions{})
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, user, users[0])
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	errGetUser := fmt.Errorf("error getting user")
	userID := uuid.New()
	user := generics.GetFakeObj[*models.User]()

	tests := map[string]struct {
		userID      string
		repoFn      func(*gomock.Controller) *repos.MockUserRepo
		errExpected error
		assertFn    func(*testing.T, *models.User)
	}{
		"happy path": {
			userID: userID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), userID).Return(user, nil)
				return repo
			},
		},
		"invalid-id": {
			userID: "foo",
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				return repo
			},
			errExpected: InvalidIDErr{},
		},
		"error": {
			userID: userID.String(),
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().Get(gomock.Any(), userID).Return(nil, errGetUser)
				return repo
			},
			errExpected: errGetUser,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &UserService{
				repo: repo,
			}
			returnedUser, err := svc.GetUser(context.Background(), test.userID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, user, returnedUser)
		})
	}
}
