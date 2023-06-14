package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestUserService_UpsertUserOrg(t *testing.T) {
	errUpsertUserOrg := fmt.Errorf("error upserting org member")
	orgID := domains.NewOrgID()
	userID := domains.NewUserID()
	userOrg := generics.GetFakeObj[*models.UserOrg]()

	tests := map[string]struct {
		orgID       string
		userID      string
		repoFn      func(*gomock.Controller) *repos.MockUserRepo
		errExpected error
		assertFn    func(*testing.T, *models.UserOrg)
	}{
		"happy path": {
			orgID:  orgID,
			userID: userID,
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().UpsertUserOrg(gomock.Any(), userID, orgID).Return(userOrg, nil)
				return repo
			},
		},
		"error": {
			orgID:  orgID,
			userID: userID,
			repoFn: func(ctl *gomock.Controller) *repos.MockUserRepo {
				repo := repos.NewMockUserRepo(ctl)
				repo.EXPECT().UpsertUserOrg(gomock.Any(), userID, orgID).Return(nil, errUpsertUserOrg)
				return repo
			},
			errExpected: errUpsertUserOrg,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			repo := test.repoFn(mockCtl)
			svc := &userService{
				log:  zaptest.NewLogger(t),
				repo: repo,
			}
			returnedUserOrg, err := svc.UpsertUserOrg(context.Background(), models.UserOrgInput{UserID: test.userID, OrgID: test.orgID})
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, returnedUserOrg)
		})
	}
}
