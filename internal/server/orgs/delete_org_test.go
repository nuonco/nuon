package orgs

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/api/internal/faker"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/request"
	orgv1 "github.com/powertoolsdev/protos/api/generated/types/org/v1"
	"github.com/stretchr/testify/assert"
)

func TestOrgService_DeleteOrg(t *testing.T) {
	errDeleteOrg := fmt.Errorf("error deleting org")

	orgReq := faker.GetFakeObj[*orgv1.DeleteOrgRequest]()

	tests := map[string]struct {
		reqFn       func(*gomock.Controller) *orgv1.DeleteOrgRequest
		repoFn      func(*gomock.Controller) *repos.MockOrgRepo
		errExpected error
	}{
		"happy path": {
			reqFn: func(ctl *gomock.Controller) *orgv1.DeleteOrgRequest {
				return orgReq
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				orgID, err := request.ParseID(orgReq.OrgId)
				assert.NoError(t, err)
				repo.EXPECT().Delete(gomock.Any(), orgID).Return(true, nil)
				return repo
			},
		},
		"invalid id": {
			reqFn: func(ctl *gomock.Controller) *orgv1.DeleteOrgRequest {
				return &orgv1.DeleteOrgRequest{
					OrgId: "invalid-id",
				}
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				return repos.NewMockOrgRepo(ctl)
			},
			errExpected: request.InvalidIDErr{},
		},
		"delete error": {
			reqFn: func(ctl *gomock.Controller) *orgv1.DeleteOrgRequest {
				return orgReq
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				orgID, err := request.ParseID(orgReq.OrgId)
				assert.NoError(t, err)
				repo.EXPECT().Delete(gomock.Any(), orgID).Return(false, errDeleteOrg)
				return repo
			},
			errExpected: errDeleteOrg,
		},
		"error deprovisioning": {
			reqFn: func(ctl *gomock.Controller) *orgv1.DeleteOrgRequest {
				return orgReq
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				orgID, err := request.ParseID(orgReq.OrgId)
				assert.NoError(t, err)
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
			svc := &server{
				repo: repo,
			}

			deleted, err := svc.DeleteOrg(context.Background(), test.reqFn(mockCtl))
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			assert.True(t, deleted.Deleted)
		})
	}
}
