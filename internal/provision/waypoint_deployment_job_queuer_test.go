package provision

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// testWaypointClientJobQueuer struct mocks a waypoint client
type testWaypointClientJobQueuer struct {
	mock.Mock
}

func (t *testWaypointClientJobQueuer) QueueJob(
	ctx context.Context,
	req *gen.QueueJobRequest,
	opts ...grpc.CallOption,
) (*gen.QueueJobResponse, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*gen.QueueJobResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func newFakeWaypointDeploymentRequest() QueueWaypointDeploymentJobRequest {
	return QueueWaypointDeploymentJobRequest{
		OrgID:        uuid.NewString(),
		AppID:        uuid.NewString(),
		InstallID:    uuid.NewString(),
		DeploymentID: uuid.NewString(),
	}
}

// Test_waypointDeployerImpl_upsertWaypointDeployment tests integrating with the waypoint deployment api
func Test_waypointDeployerImpl_upsertWaypointDeployment(t *testing.T) {
	errDeployment := fmt.Errorf("error upserting deployment")
	req := newFakeWaypointDeploymentRequest()

	tests := map[string]struct {
		waypointHcl []byte
		clientFn    func() waypointClientJobQueuer
		assertFn    func(*testing.T, waypointClientJobQueuer, string)
		expectedErr error
	}{
		"happy path": {
			clientFn: func() waypointClientJobQueuer {
				client := &testWaypointClientJobQueuer{}
				client.On("QueueJob", mock.Anything, mock.Anything, mock.Anything).Return(&gen.QueueJobResponse{}, nil)
				return client
			},
			expectedErr: nil,
			assertFn: func(t *testing.T, client waypointClientJobQueuer, _ string) {
				obj := client.(*testWaypointClientJobQueuer)
				obj.AssertNumberOfCalls(t, "QueueJob", 1)

				wpReq := obj.Calls[0].Arguments[1].(*gen.QueueJobRequest)
				assert.NotNil(t, wpReq)
				assert.Equal(t, fmt.Sprintf("%s-%s", req.InstallID, req.DeploymentID), wpReq.Job.SingletonId)

				// assert runner
				assert.Equal(t, req.InstallID, wpReq.Job.OndemandRunner.Name)

				// assert project
				assert.Equal(t, req.InstallID, wpReq.Job.Application.Project)
				assert.Equal(t, req.ComponentName, wpReq.Job.Application.Application)

				// assert labels
				assert.Equal(t, req.InstallID, wpReq.Job.Labels["install_id"])
				assert.Equal(t, req.DeploymentID, wpReq.Job.Labels["deployment_id"])

				// TODO(jm): test artifacts, waypointhcl and variables once implemented
			},
		},
		"client err": {
			clientFn: func() waypointClientJobQueuer {
				client := &testWaypointClientJobQueuer{}
				client.On("QueueJob", mock.Anything, mock.Anything, mock.Anything).Return(nil, errDeployment)
				return client
			},
			expectedErr: errDeployment,
			assertFn: func(t *testing.T, client waypointClientJobQueuer, _ string) {
				obj := client.(*testWaypointClientJobQueuer)
				obj.AssertNumberOfCalls(t, "QueueJob", 1)

				wpReq := obj.Calls[0].Arguments[1].(*gen.QueueJobRequest)
				assert.NotNil(t, wpReq)
				// assert.Equal(t, wpReq.Deployment.Application.Project, req.InstallID)
				// assert.Equal(t, wpReq.Deployment.Application.Application, req.ComponentName)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			w := &waypointDeploymentJobQueuerImpl{}
			client := test.clientFn()

			jobID, err := w.queueWaypointDeploymentJob(context.Background(), client, req, test.waypointHcl, nil)
			if test.expectedErr != nil {
				assert.ErrorContains(t, err, test.expectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			test.assertFn(t, client, jobID)
		})
	}
}

// testS3ClientObjectGetter struct mocks a waypoint client
type testS3ClientObjectGetter struct {
	mock.Mock
}

func (t *testS3ClientObjectGetter) GetObject(
	ctx context.Context,
	req *s3.GetObjectInput,
	opts ...func(*s3.Options),
) (*s3.GetObjectOutput, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

func Test_waypointDeploymentJobQueuerImpl_getWaypointHcl(t *testing.T) {
	errGetObject := fmt.Errorf("error getting object")
	req := newFakeWaypointDeploymentRequest()

	tests := map[string]struct {
		clientFn    func() s3ClientObjectGetter
		assertFn    func(*testing.T, s3ClientObjectGetter, []byte)
		expectedErr error
	}{
		"happy path": {
			clientFn: func() s3ClientObjectGetter {
				client := &testS3ClientObjectGetter{}
				r := io.NopCloser(strings.NewReader("cfg")) // r type is io.ReadCloser
				client.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{
					Body: r,
				}, nil)
				return client
			},
			expectedErr: nil,
			assertFn: func(t *testing.T, client s3ClientObjectGetter, byts []byte) {
				obj := client.(*testS3ClientObjectGetter)
				obj.AssertNumberOfCalls(t, "GetObject", 1)

				s3Req := obj.Calls[0].Arguments[1].(*s3.GetObjectInput)
				assert.NotNil(t, s3Req)
				assert.Equal(t, req.BucketName, *s3Req.Bucket)
				assert.Equal(t, fmt.Sprintf("%s/deploy.hcl", req.BucketPrefix), *s3Req.Key)
				assert.Equal(t, []byte("cfg"), byts)
			},
		},
		"client err": {
			clientFn: func() s3ClientObjectGetter {
				client := &testS3ClientObjectGetter{}
				client.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(nil, errGetObject)
				return client
			},
			expectedErr: errGetObject,
			assertFn: func(t *testing.T, client s3ClientObjectGetter, _ []byte) {
				obj := client.(*testS3ClientObjectGetter)
				obj.AssertNumberOfCalls(t, "GetObject", 1)

				s3Req := obj.Calls[0].Arguments[1].(*s3.GetObjectInput)
				assert.NotNil(t, s3Req)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			w := &waypointDeploymentJobQueuerImpl{}
			client := test.clientFn()

			waypointHcl, err := w.getWaypointHcl(context.Background(), client, req)
			if test.expectedErr != nil {
				assert.ErrorContains(t, err, test.expectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			test.assertFn(t, client, waypointHcl)
		})
	}
}
