package services

import (
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/orgs-api/internal/repos/s3"
	"github.com/powertoolsdev/orgs-api/internal/repos/waypoint"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	mockCtl := gomock.NewController(t)
	s3Repo := s3.NewMockRepo(mockCtl)
	waypointRepo := waypoint.NewMockRepo(mockCtl)
	workflowsRepo := workflows.NewMockRepo(mockCtl)

	tests := map[string]struct {
		optFns      func() []serviceOption
		assertFn    func(*testing.T, *service)
		errExpected error
	}{
		"happy path": {
			optFns: func() []serviceOption {
				return []serviceOption{
					WithS3Repo(s3Repo),
					WithWaypointRepo(waypointRepo),
					WithWorkflowsRepo(workflowsRepo),
				}
			},
			assertFn: func(t *testing.T, srv *service) {
				assert.Equal(t, s3Repo, srv.S3Repo)
				assert.Equal(t, waypointRepo, srv.WaypointRepo)
				assert.Equal(t, workflowsRepo, srv.WorkflowsRepo)
			},
		},
		"missing s3 repo": {
			optFns: func() []serviceOption {
				return []serviceOption{
					WithWaypointRepo(waypointRepo),
					WithWorkflowsRepo(workflowsRepo),
				}
			},
			errExpected: fmt.Errorf("service.S3Repo"),
		},
		"missing waypoint repo": {
			optFns: func() []serviceOption {
				return []serviceOption{
					WithS3Repo(s3Repo),
					WithWorkflowsRepo(workflowsRepo),
				}
			},
			errExpected: fmt.Errorf("service.WaypointRepo"),
		},
		"missing workflows repo": {
			optFns: func() []serviceOption {
				return []serviceOption{
					WithS3Repo(s3Repo),
					WithWaypointRepo(waypointRepo),
				}
			},
			errExpected: fmt.Errorf("service.WorkflowsRepo"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			opts := test.optFns()
			srv, err := New(opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, srv)
		})
	}
}
