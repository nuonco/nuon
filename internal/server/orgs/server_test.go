package orgs

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/api/internal/faker"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/request"
	orgV1 "github.com/powertoolsdev/protos/api/generated/types/org/v1"
	"github.com/stretchr/testify/assert"
)

func TestOrgService_GetOrg(t *testing.T) {
	org := faker.GetFakeObj[*models.Org]()
	orgReq := faker.GetFakeObj[*orgV1.GetOrgRequest]()

	tests := map[string]struct {
		reqFn       func(*gomock.Controller) *orgV1.GetOrgRequest
		repoFn      func(*gomock.Controller) *repos.MockOrgRepo
		errExpected error
		assertFn    func(*testing.T, *orgV1.GetOrgResponse, *repos.MockOrgRepo)
	}{
		"happy path": {
			reqFn: func(ctl *gomock.Controller) *orgV1.GetOrgRequest {
				return orgReq
			},
			repoFn: func(ctl *gomock.Controller) *repos.MockOrgRepo {
				repo := repos.NewMockOrgRepo(ctl)
				reqUUID, err := request.ParseID(orgReq.OrgId)
				assert.NoError(t, err)
				repo.EXPECT().Get(gomock.Any(), reqUUID).Return(org, nil)
				return repo
			},
			assertFn: func(t *testing.T, resp *orgV1.GetOrgResponse, mg *repos.MockOrgRepo) {
				assert.NotNil(t, resp.Org)
				assert.Equal(t, org.ID.String(), resp.Org.Id)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			req := test.reqFn(mockCtl)
			repo := test.repoFn(mockCtl)

			srv := &server{
				repo: repo,
			}

			resp, err := srv.GetOrg(context.Background(), req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			test.assertFn(t, resp, repo)
		})
	}
}
