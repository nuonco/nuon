package services

import (
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/orgs-api/internal/repos/s3"
	"github.com/powertoolsdev/orgs-api/internal/repos/waypoint"
	"gotest.tools/assert"
)

func TestNew(t *testing.T) {
	mockCtl := gomock.NewController(t)
	s3Repo := s3.NewMockRepo(mockCtl)
	waypointRepo := waypoint.NewMockRepo(mockCtl)

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
				}
			},
			assertFn: func(t *testing.T, srv *service) {
				assert.Equal(t, s3Repo, srv.S3Repo)
				assert.Equal(t, waypointRepo, srv.WaypointRepo)
			},
		},
		"missing s3 repo": {
			optFns: func() []serviceOption {
				return []serviceOption{
					WithWaypointRepo(waypointRepo),
				}
			},
			errExpected: fmt.Errorf("service.S3Repo"),
		},
		"missing waypoint repo": {
			optFns: func() []serviceOption {
				return []serviceOption{
					WithS3Repo(s3Repo),
				}
			},
			errExpected: fmt.Errorf("service.WaypointRepo"),
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

			test.assertFn(t, srv)
		})
	}
}
