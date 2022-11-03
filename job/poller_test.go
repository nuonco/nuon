package job

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	status "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
)

type testWaypointClientJobPoller struct {
	mock.Mock
}

func (t *testWaypointClientJobPoller) GetJobStream(
	ctx context.Context,
	req *gen.GetJobStreamRequest,
	opts ...grpc.CallOption,
) (gen.Waypoint_GetJobStreamClient, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(gen.Waypoint_GetJobStreamClient), args.Error(1)
	}

	return nil, args.Error(1)
}

type testWaypointClientJobStreamReceiver struct {
	mock.Mock

	// NOTE(jm): we embed the gen interface here so we don't have to implement the set of mock methods for this type
	// that we're not actually using when using this value as a return value during mocking
	gen.Waypoint_GetJobStreamClient
}

func (t *testWaypointClientJobStreamReceiver) Recv() (*gen.GetJobStreamResponse, error) {
	args := t.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*gen.GetJobStreamResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func Test_waypointDeploymentJobPollerImpl_getWaypointDeploymentJobStream(t *testing.T) {
	errGetJobStream := fmt.Errorf("err get job stream")

	tests := map[string]struct {
		clientFn    func() waypointClientJobPoller
		jobID       string
		assertFn    func(waypointClientJobPoller)
		errExpected error
	}{
		"happy path": {
			clientFn: func() waypointClientJobPoller {
				obj := &testWaypointClientJobPoller{}
				recvr := &testWaypointClientJobStreamReceiver{}
				obj.On("GetJobStream", mock.Anything, mock.Anything, mock.Anything).Return(recvr, nil)
				return obj
			},
			jobID:       "job-id-foobar",
			errExpected: nil,
			assertFn: func(client waypointClientJobPoller) {
				obj := client.(*testWaypointClientJobPoller)
				obj.AssertNumberOfCalls(t, "GetJobStream", 1)

				jsReq := obj.Calls[0].Arguments[1].(*gen.GetJobStreamRequest)
				assert.Equal(t, "job-id-foobar", jsReq.JobId)
				assert.NotNil(t, jsReq)
			},
		},
		"err returned": {
			clientFn: func() waypointClientJobPoller {
				obj := &testWaypointClientJobPoller{}
				obj.On("GetJobStream", mock.Anything, mock.Anything, mock.Anything).Return(nil, errGetJobStream)
				return obj
			},
			jobID:       "job-id-foobar",
			errExpected: errGetJobStream,
			assertFn: func(client waypointClientJobPoller) {
				obj := client.(*testWaypointClientJobPoller)
				obj.AssertNumberOfCalls(t, "GetJobStream", 1)

				jsReq := obj.Calls[0].Arguments[1].(*gen.GetJobStreamRequest)
				assert.Equal(t, "job-id-foobar", jsReq.JobId)
				assert.NotNil(t, jsReq)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.clientFn()
			impl := waypointDeploymentJobPollerImpl{}
			streamClient, err := impl.getWaypointDeploymentJobStream(context.Background(), client, test.jobID)

			if test.errExpected != nil {
				assert.NotNil(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, streamClient)
			}
		})
	}
}

type testWaypointJobEventWriter struct {
	mock.Mock
}

func (t *testWaypointJobEventWriter) Write(ev WaypointJobEvent) error {
	args := t.Called(ev)
	return args.Error(0)
}

func Test_waypointDeploymentJobPollerImpl_consumeWaypointDeploymentJobStream(t *testing.T) {
	errConsumeWaypointDeploymentJobStream := fmt.Errorf("err unable to consume waypoint job stream")
	jobID := uuid.NewString()

	tests := map[string]struct {
		recvrFn     func(func()) waypointClientJobStreamReceiver
		errExpected error
		writerFn    func() EventWriter
		jobID       string
		assertFn    func(*testing.T, waypointClientJobStreamReceiver, EventWriter)
	}{
		"happy path": {
			recvrFn: func(cancelFn func()) waypointClientJobStreamReceiver {
				obj := &testWaypointClientJobStreamReceiver{}
				obj.On("Recv").Return(&gen.GetJobStreamResponse{
					Event: &gen.GetJobStreamResponse_Open_{
						Open: &gen.GetJobStreamResponse_Open{},
					},
				}, nil).Run(func(mock.Arguments) {
					cancelFn()
				})
				return obj
			},
			errExpected: context.Canceled,
			writerFn: func() EventWriter {
				obj := &testWaypointJobEventWriter{}
				obj.On("Write", mock.Anything).Return(nil)
				return obj
			},
			jobID: jobID,
			assertFn: func(t *testing.T, client waypointClientJobStreamReceiver, wjew EventWriter) {
				recvr := client.(*testWaypointClientJobStreamReceiver)
				recvr.AssertNumberOfCalls(t, "Recv", 1)

				writer := wjew.(*testWaypointJobEventWriter)
				writer.AssertNumberOfCalls(t, "Write", 1)
			},
		},
		"error-receiving": {
			recvrFn: func(cancelFn func()) waypointClientJobStreamReceiver {
				obj := &testWaypointClientJobStreamReceiver{}
				obj.On("Recv").Return(nil, errConsumeWaypointDeploymentJobStream)
				return obj
			},
			errExpected: errConsumeWaypointDeploymentJobStream,
			writerFn: func() EventWriter {
				obj := &testWaypointJobEventWriter{}
				obj.On("Write", mock.Anything).Return(nil)
				return obj
			},
			jobID: jobID,
			assertFn: func(t *testing.T, client waypointClientJobStreamReceiver, wjew EventWriter) {
				recvr := client.(*testWaypointClientJobStreamReceiver)
				recvr.AssertNumberOfCalls(t, "Recv", 1)

				writer := wjew.(*testWaypointJobEventWriter)
				writer.AssertNumberOfCalls(t, "Write", 0)
			},
		},
		"stream-error-event": {
			recvrFn: func(cancelFn func()) waypointClientJobStreamReceiver {
				obj := &testWaypointClientJobStreamReceiver{}
				obj.On("Recv").Return(&gen.GetJobStreamResponse{
					Event: &gen.GetJobStreamResponse_Error_{
						Error: &gen.GetJobStreamResponse_Error{
							Error: &status.Status{
								Code:    400,
								Message: "error",
							},
						},
					},
				}, nil)
				return obj
			},
			errExpected: errWaypointJobStream,
			writerFn: func() EventWriter {
				obj := &testWaypointJobEventWriter{}
				obj.On("Write", mock.Anything).Return(nil)
				return obj
			},
			jobID: jobID,
			assertFn: func(t *testing.T, client waypointClientJobStreamReceiver, wjew EventWriter) {
				recvr := client.(*testWaypointClientJobStreamReceiver)
				recvr.AssertNumberOfCalls(t, "Recv", 1)

				writer := wjew.(*testWaypointJobEventWriter)
				writer.AssertNumberOfCalls(t, "Write", 0)
			},
		},
		"error-writing": {
			recvrFn: func(cancelFn func()) waypointClientJobStreamReceiver {
				obj := &testWaypointClientJobStreamReceiver{}
				obj.On("Recv").Return(&gen.GetJobStreamResponse{
					Event: &gen.GetJobStreamResponse_Open_{
						Open: &gen.GetJobStreamResponse_Open{},
					},
				}, nil)
				return obj
			},
			errExpected: errConsumeWaypointDeploymentJobStream,
			writerFn: func() EventWriter {
				obj := &testWaypointJobEventWriter{}
				obj.On("Write", mock.Anything).Return(errConsumeWaypointDeploymentJobStream)
				return obj
			},
			jobID: jobID,
			assertFn: func(t *testing.T, client waypointClientJobStreamReceiver, wjew EventWriter) {
				recvr := client.(*testWaypointClientJobStreamReceiver)
				recvr.AssertNumberOfCalls(t, "Recv", 1)
			},
		},
		"job-failure": {
			recvrFn: func(cancelFn func()) waypointClientJobStreamReceiver {
				obj := &testWaypointClientJobStreamReceiver{}
				obj.On("Recv").Return(&gen.GetJobStreamResponse{
					Event: &gen.GetJobStreamResponse_Complete_{
						Complete: &gen.GetJobStreamResponse_Complete{
							Error: &status.Status{
								Message: "failure",
							},
							Result: &gen.Job_Result{},
						},
					},
				}, nil)
				return obj
			},
			errExpected: errWaypointJobFailed,
			writerFn: func() EventWriter {
				obj := &testWaypointJobEventWriter{}
				obj.On("Write", mock.Anything).Return(nil)
				return obj
			},
			jobID: jobID,
			assertFn: func(t *testing.T, client waypointClientJobStreamReceiver, wjew EventWriter) {
				recvr := client.(*testWaypointClientJobStreamReceiver)
				recvr.AssertNumberOfCalls(t, "Recv", 1)

				twjew := wjew.(*testWaypointJobEventWriter)
				ev := twjew.Calls[0].Arguments[0].(*WaypointJobEvent)
				assert.Equal(t, ev.Type, waypointJobEventTypeComplete)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancelFn := context.WithCancel(ctx)

			obj := &waypointDeploymentJobPollerImpl{}
			client := test.recvrFn(cancelFn)
			writer := test.writerFn()

			err := obj.consumeWaypointDeploymentJobStream(ctx, test.jobID, client, writer)
			if test.errExpected != nil {
				assert.NotNil(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
