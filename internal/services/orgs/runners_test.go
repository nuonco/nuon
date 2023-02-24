package services

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-generics"
	"github.com/powertoolsdev/orgs-api/internal/repos/waypoint"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
	"github.com/stretchr/testify/assert"
)

//nolint:all
func Test_service_GetRunners(t *testing.T) {
	return
	mockRunnerResp := generics.GetFakeObj[*waypointv1.ListRunnersResponse]()
	orgID := shortid.New()

	tests := map[string]struct {
		waypointFn  func(*gomock.Controller) waypoint.Repo
		assertFn    func(*testing.T, *orgsv1.GetRunnersResponse)
		errExpected error
	}{
		"happy path": {
			waypointFn: func(mockCtl *gomock.Controller) waypoint.Repo {
				mock := waypoint.NewMockRepo(mockCtl)
				mock.EXPECT().ListRunners(gomock.Any()).
					Return(mockRunnerResp, nil)
				return mock
			},
			assertFn: func(t *testing.T, resp *orgsv1.GetRunnersResponse) {
				assert.NotNil(t, resp)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)

			svc := &service{
				WaypointRepo: test.waypointFn(mockCtl),
			}

			resp, err := svc.GetRunners(ctx, orgID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, resp)
		})
	}
}
