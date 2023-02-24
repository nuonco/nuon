package waypoint

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"

	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func Test_repo_ListRunners(t *testing.T) {
	errListRunners := fmt.Errorf("error listing runners")
	listRunnersResponse := &waypointv1.ListRunnersResponse{}

	tests := map[string]struct {
		clientGetter func(*gomock.Controller) clientGetter
		assertFn     func(*testing.T, *waypointv1.ListRunnersResponse)
		errExpected  error
	}{
		"happy path": {
			clientGetter: func(mockCtl *gomock.Controller) clientGetter {
				mock := NewMockwaypointClient(mockCtl)
				req := &waypointv1.ListRunnersRequest{}
				mock.EXPECT().ListRunners(gomock.Any(), req, gomock.Any()).
					Return(listRunnersResponse, nil)

				return func(context.Context) (waypointClient, error) {
					return mock, nil
				}
			},
			assertFn: func(t *testing.T, resp *waypointv1.ListRunnersResponse) {
				assert.True(t, proto.Equal(resp, listRunnersResponse))
			},
		},
		"unable to get client err": {
			clientGetter: func(mockCtl *gomock.Controller) clientGetter {
				return func(context.Context) (waypointClient, error) {
					return nil, errListRunners
				}
			},
			errExpected: errListRunners,
		},
		"client error": {
			clientGetter: func(mockCtl *gomock.Controller) clientGetter {
				mock := NewMockwaypointClient(mockCtl)
				mock.EXPECT().ListRunners(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errListRunners)

				return func(context.Context) (waypointClient, error) {
					return mock, nil
				}
			},
			errExpected: errListRunners,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)
			repo := &repo{
				ClientGetter: test.clientGetter(mockCtl),
			}

			resp, err := repo.ListRunners(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, resp)
		})
	}
}
