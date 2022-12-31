package waypoint

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/orgs-api/internal/orgcontext"
	"github.com/stretchr/testify/assert"
)

func Test_repo_getClient(t *testing.T) {
	orgCtx := getFakeOrgContext()
	parentMockCtl := gomock.NewController(t)
	mockWaypointClient := NewMockwaypointClient(parentMockCtl)
	errGetClient := fmt.Errorf("error getting client")

	tests := map[string]struct {
		ctxGetterFn        orgContextGetter
		waypointProviderFn func(*gomock.Controller) waypointClientProvider
		assertFn           func(*testing.T, waypointClient)
		errExpected        error
	}{
		"happy path": {
			ctxGetterFn: func(ctx context.Context) (*orgcontext.Context, error) {
				return orgCtx, nil
			},
			waypointProviderFn: func(mockCtl *gomock.Controller) waypointClientProvider {
				mock := NewMockwaypointClientProvider(mockCtl)
				mock.EXPECT().GetOrgWaypointClient(gomock.Any(),
					orgCtx.WaypointServer.SecretNamespace,
					orgCtx.OrgID,
					orgCtx.WaypointServer.Address).Return(mockWaypointClient, nil)
				return mock
			},
			assertFn: func(t *testing.T, client waypointClient) {
				assert.Equal(t, mockWaypointClient, client)
			},
		},
		"org context error": {
			ctxGetterFn: func(ctx context.Context) (*orgcontext.Context, error) {
				return nil, errGetClient
			},
			waypointProviderFn: func(mockCtl *gomock.Controller) waypointClientProvider {
				mock := NewMockwaypointClientProvider(mockCtl)
				return mock
			},
			errExpected: errGetClient,
		},
		"client error": {
			ctxGetterFn: func(ctx context.Context) (*orgcontext.Context, error) {
				return orgCtx, nil
			},
			waypointProviderFn: func(mockCtl *gomock.Controller) waypointClientProvider {
				mock := NewMockwaypointClientProvider(mockCtl)
				mock.EXPECT().GetOrgWaypointClient(gomock.Any(),
					orgCtx.WaypointServer.SecretNamespace,
					orgCtx.OrgID,
					orgCtx.WaypointServer.Address).Return(nil, errGetClient)
				return mock
			},
			errExpected: errGetClient,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)
			repo := &repo{
				CtxGetter:              test.ctxGetterFn,
				WaypointClientProvider: test.waypointProviderFn(mockCtl),
			}

			client, err := repo.getClient(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, client)
		})
	}
}
