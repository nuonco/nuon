package build

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type testWaypointClientApplicationUpserter struct {
	mock.Mock
}

func (t *testWaypointClientApplicationUpserter) UpsertApplication(ctx context.Context, req *gen.UpsertApplicationRequest, opts ...grpc.CallOption) (*gen.UpsertApplicationResponse, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*gen.UpsertApplicationResponse), args.Error(1)
	}

	return nil, args.Error(1)
}
func TestUpsertWaypointApplication_validation(t *testing.T) {
	tests := map[string]struct {
		reqFn       func() UpsertWaypointApplicationRequest
		errExpected error
	}{
		"happy path": {
			reqFn: getFakeObj[UpsertWaypointApplicationRequest],
		},
		"missing-org-id": {
			reqFn: func() UpsertWaypointApplicationRequest {
				req := getFakeObj[UpsertWaypointApplicationRequest]()
				req.OrgID = ""
				return req
			},
			errExpected: fmt.Errorf("UpsertWaypointApplicationRequest.OrgID"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := test.reqFn()
			err := req.validate()
			if test.errExpected == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorContains(t, err, test.errExpected.Error())
			}
		})
	}
}

func TestUpsertWaypointApplication_upsertApplication(t *testing.T) {
	req := getFakeObj[UpsertWaypointApplicationRequest]()
	testErr := fmt.Errorf("test-error")

	tests := map[string]struct {
		clientFn    func() waypointClientApplicationUpserter
		assertFn    func(*testing.T, waypointClientApplicationUpserter)
		errExpected error
	}{
		"happy path": {
			clientFn: func() waypointClientApplicationUpserter {
				client := &testWaypointClientApplicationUpserter{}
				client.On("UpsertApplication", mock.Anything, mock.Anything, mock.Anything).Return(&gen.UpsertApplicationResponse{}, nil)
				return client
			},
			assertFn: func(t *testing.T, client waypointClientApplicationUpserter) {
				obj := client.(*testWaypointClientApplicationUpserter)
				obj.AssertNumberOfCalls(t, "UpsertApplication", 1)

				wpReq := obj.Calls[0].Arguments[1].(*gen.UpsertApplicationRequest)

				assert.Equal(t, wpReq.Name, req.Component.Name)
				assert.Equal(t, wpReq.Project.Project, req.OrgID)
			},
		},
		"error": {
			clientFn: func() waypointClientApplicationUpserter {
				client := &testWaypointClientApplicationUpserter{}
				client.On("UpsertApplication", mock.Anything, mock.Anything, mock.Anything).Return(nil, testErr)
				return client
			},
			errExpected: testErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			apper := wpApplicationUpserter{}
			client := test.clientFn()

			err := apper.upsertWaypointApplication(context.Background(), client, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
		})
	}
}
