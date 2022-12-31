package services

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/google/uuid"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-generics"
	"github.com/powertoolsdev/orgs-api/internal/repos/waypoint"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
	"github.com/stretchr/testify/assert"
)

func Test_service_GetInfo(t *testing.T) {
	//errGetInfo := fmt.Errorf("error getting info")
	orgID, err := shortid.ParseUUID(uuid.New())
	assert.NoError(t, err)

	mockRunnerResp := &waypointv1.Runner{Id: orgID}
	mockVersionInfoResp := generics.GetFakeObj[*waypointv1.GetVersionInfoResponse]()
	mockWorkflowResp := generics.GetFakeObj[*sharedv1.Response]()
	mockWorkflowResp.Response = &sharedv1.ResponseRef{
		Response: generics.GetFakeObj[*sharedv1.ResponseRef_OrgSignup](),
	}

	tests := map[string]struct {
		waypointFn  func(*gomock.Controller) waypoint.Repo
		workflowsFn func(*gomock.Controller) workflows.Repo
		assertFn    func(*testing.T, *orgsv1.GetInfoResponse)
		errExpected error
	}{
		"happy path": {
			waypointFn: func(mockCtl *gomock.Controller) waypoint.Repo {
				mock := waypoint.NewMockRepo(mockCtl)
				mock.EXPECT().GetRunner(gomock.Any(), orgID).
					Return(mockRunnerResp, nil)
				mock.EXPECT().GetVersionInfo(gomock.Any()).
					Return(mockVersionInfoResp, nil)
				return mock
			},
			workflowsFn: func(mockCtl *gomock.Controller) workflows.Repo {
				mock := workflows.NewMockRepo(mockCtl)
				mock.EXPECT().GetOrgProvisionResponse(gomock.Any()).
					Return(mockWorkflowResp, nil)
				return mock
			},
			assertFn: func(t *testing.T, resp *orgsv1.GetInfoResponse) {
				assert.NotNil(t, resp)

				// TODO(jm): turn this back on once we've gotten the correct proto format
				// assert.NoError(t, resp.Validate())
			},
		},
		//"error getting runner": {
		//waypointFn: func(mockCtl *gomock.Controller) waypoint.Repo {
		//mock := waypoint.NewMockRepo(mockCtl)
		//return mock
		//},
		//workflowsFn: func(mockCtl *gomock.Controller) workflows.Repo {
		//mock := workflows.NewMockRepo(mockCtl)
		//return mock
		//},
		//errExpected: errGetInfo,
		//},
		//"error getting workflow": {
		//waypointFn: func(mockCtl *gomock.Controller) waypoint.Repo {
		//mock := waypoint.NewMockRepo(mockCtl)
		//return mock
		//},
		//workflowsFn: func(mockCtl *gomock.Controller) workflows.Repo {
		//mock := workflows.NewMockRepo(mockCtl)
		//return mock
		//},
		//errExpected: errGetInfo,
		//},
		//"error getting server info": {
		//waypointFn: func(mockCtl *gomock.Controller) waypoint.Repo {
		//mock := waypoint.NewMockRepo(mockCtl)
		//return mock
		//},
		//workflowsFn: func(mockCtl *gomock.Controller) workflows.Repo {
		//mock := workflows.NewMockRepo(mockCtl)
		//return mock
		//},
		//errExpected: errGetInfo,
		//},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)
			svc := &service{
				WorkflowsRepo: test.workflowsFn(mockCtl),
				WaypointRepo:  test.waypointFn(mockCtl),
			}

			resp, err := svc.GetInfo(ctx, orgID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, resp)
		})
	}
}
