package orgs

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/powertoolsdev/orgs-api/internal/orgcontext"
	orgsservice "github.com/powertoolsdev/orgs-api/internal/services/orgs"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func Test_server_GetInfo(t *testing.T) {
	orgCtx := getFakeOrgContext()
	orgID := uuid.NewString()
	getInfoResp := &orgsv1.GetInfoResponse{}

	tests := map[string]struct {
		serverFn    func(*testing.T, *gomock.Controller) *server
		reqFn       func(*testing.T) *connect.Request[orgsv1.GetInfoRequest]
		assertFn    func(*testing.T, *connect.Response[orgsv1.GetInfoResponse])
		errExpected error
	}{
		"happy path": {
			serverFn: func(t *testing.T, mockCtl *gomock.Controller) *server {
				mockOrgCtx := orgcontext.NewMockProvider(mockCtl)
				mockOrgCtx.EXPECT().SetContext(gomock.Any(), orgID).DoAndReturn(
					func(ctx context.Context, _ string) (context.Context, error) {
						ctx = context.WithValue(ctx, orgcontext.Key{}, orgCtx)
						return ctx, nil
					},
				)
				mockSvc := orgsservice.NewMockService(mockCtl)
				mockSvc.EXPECT().GetInfo(gomock.Any(), orgID).Return(getInfoResp, nil)

				srv, err := New(WithContextProvider(mockOrgCtx), WithService(mockSvc))
				assert.NoError(t, err)
				return srv
			},
			reqFn: func(t *testing.T) *connect.Request[orgsv1.GetInfoRequest] {
				return connect.NewRequest(&orgsv1.GetInfoRequest{
					OrgId: orgID,
				})
			},
			assertFn: func(t *testing.T, res *connect.Response[orgsv1.GetInfoResponse]) {
				assert.NotNil(t, res)
				assert.True(t, proto.Equal(res.Msg, getInfoResp))
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)
			server := test.serverFn(t, mockCtl)
			req := test.reqFn(t)

			res, err := server.GetInfo(ctx, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, res)
		})
	}
}
