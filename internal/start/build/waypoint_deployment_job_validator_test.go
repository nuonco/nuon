package build

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// testWaypointClientJobValidator struct mocks a waypoint client
type testWaypointClientJobValidator struct {
	mock.Mock
}

func (t *testWaypointClientJobValidator) ValidateJob(
	ctx context.Context,
	req *gen.ValidateJobRequest,
	opts ...grpc.CallOption,
) (*gen.ValidateJobResponse, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*gen.ValidateJobResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

// Test_waypointDeployerImpl_validateWaypointDeployment tests integrating with the waypoint deployment api
func Test_waypointDeployerImpl_validateWaypointDeployment(t *testing.T) {
	errValidate := fmt.Errorf("error validating deployment")
	jobID := uuid.NewString()

	tests := map[string]struct {
		waypointHcl []byte
		clientFn    func() waypointClientJobValidator
		assertFn    func(*testing.T, waypointClientJobValidator, string)
		expectedErr error
	}{
		"happy path": {
			clientFn: func() waypointClientJobValidator {
				client := &testWaypointClientJobValidator{}
				client.On("ValidateJob", mock.Anything, mock.Anything, mock.Anything).Return(&gen.ValidateJobResponse{}, nil)
				return client
			},
			expectedErr: nil,
			assertFn: func(t *testing.T, client waypointClientJobValidator, _ string) {
				obj := client.(*testWaypointClientJobValidator)
				obj.AssertNumberOfCalls(t, "ValidateJob", 1)

				wpReq := obj.Calls[0].Arguments[1].(*gen.ValidateJobRequest)
				assert.NotNil(t, wpReq)
				// assert.Equal(t, wpReq.Deployment.Application.Project, req.InstallID)
				// assert.Equal(t, wpReq.Deployment.Application.Application, req.ComponentName)
			},
		},
		"validate job client err": {
			clientFn: func() waypointClientJobValidator {
				client := &testWaypointClientJobValidator{}

				client.On("ValidateJob", mock.Anything, mock.Anything, mock.Anything).Return(nil, errValidate)
				return client
			},
			expectedErr: errValidate,
			assertFn: func(t *testing.T, client waypointClientJobValidator, _ string) {
				obj := client.(*testWaypointClientJobValidator)
				obj.AssertNumberOfCalls(t, "ValidateJob", 1)

				wpReq := obj.Calls[0].Arguments[1].(*gen.ValidateJobRequest)
				assert.NotNil(t, wpReq)
				// assert.Equal(t, wpReq.Deployment.Application.Project, req.InstallID)
				// assert.Equal(t, wpReq.Deployment.Application.Application, req.ComponentName)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			w := &waypointDeploymentJobValidatorImpl{}
			client := test.clientFn()

			err := w.validateWaypointDeploymentJob(context.Background(), client, jobID)
			if test.expectedErr != nil {
				assert.ErrorContains(t, err, test.expectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			test.assertFn(t, client, jobID)
		})
	}
}
