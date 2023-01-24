package poll

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/workers-executors/internal/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	status "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
)

func TestNew(t *testing.T) {
	t.Parallel()
	v := validator.New()

	_ = v

	tests := map[string]struct {
		v           *validator.Validate
		opts        []pollerOption
		errExpected error
	}{
		"happy path": {
			v: v,
			opts: []pollerOption{
				WithClient(&mockJobStreamGetter{}),
				WithJobID(uuid.NewString()),
				WithWriter(&mockEventWriter{}),
			},
		},
		"missing validator": {
			v: nil,
			opts: []pollerOption{
				WithClient(&mockJobStreamGetter{}),
				WithJobID(uuid.NewString()),
				WithWriter(&mockEventWriter{}),
			},
			errExpected: fmt.Errorf("validator is nil"),
		},
		"missing client": {
			v: v,
			opts: []pollerOption{
				WithJobID(uuid.NewString()),
				WithWriter(&mockEventWriter{}),
			},
			errExpected: fmt.Errorf("validation for 'Client' failed on the 'required' tag"),
		},
		"missing job id": {
			v: v,
			opts: []pollerOption{
				WithClient(&mockJobStreamGetter{}),
				WithWriter(&mockEventWriter{}),
			},
			errExpected: fmt.Errorf("validation for 'JobID' failed on the 'required' tag"),
		},
		"missing writer": {
			v: v,
			opts: []pollerOption{
				WithClient(&mockJobStreamGetter{}),
				WithJobID(uuid.NewString()),
			},
			errExpected: fmt.Errorf("validation for 'Writer' failed on the 'required' tag"),
		},
	}

	for name, test := range tests {
		name := name
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p, err := New(test.v, test.opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, p)
		})
	}
}

type mockJobStreamGetter struct {
	mock.Mock
}

func (t *mockJobStreamGetter) GetJobStream(
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

var _ jobStreamGetter = (*mockJobStreamGetter)(nil)

type mockReceiver struct {
	mock.Mock

	// NOTE(jm): we embed the gen interface here so we don't have to implement the set of mock methods for this type
	// that we're not actually using when using this value as a return value during mocking
	gen.Waypoint_GetJobStreamClient
}

func (t *mockReceiver) Recv() (*gen.GetJobStreamResponse, error) {
	args := t.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*gen.GetJobStreamResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

type mockEventWriter struct {
	mock.Mock
}

func (t *mockEventWriter) Write(ev event.WaypointJobEvent) error {
	args := t.Called(ev)
	return args.Error(0)
}

var _ eventWriter = (*mockEventWriter)(nil)

func TestPoller_Poll(t *testing.T) {
	t.Parallel()

	v := validator.New()

	errConsumeWaypointDeploymentJobStream := fmt.Errorf("err unable to consume waypoint job stream")
	// errGetJobStream := fmt.Errorf("err get job stream")

	tests := map[string]struct {
		jobStreamGetter func(*testing.T, func()) *mockJobStreamGetter
		writer          func(*testing.T) *mockEventWriter
		errExpected     error
	}{
		"happy path": {
			jobStreamGetter: func(t *testing.T, cnclFn func()) *mockJobStreamGetter {
				recvr := &mockReceiver{}
				recvr.
					On("Recv").
					Return(&gen.GetJobStreamResponse{
						Event: &gen.GetJobStreamResponse_Open_{
							Open: &gen.GetJobStreamResponse_Open{},
						},
					}, nil).
					Run(func(mock.Arguments) { cnclFn() })

				obj := &mockJobStreamGetter{}
				obj.
					On(
						"GetJobStream",
						mock.Anything,
						mock.MatchedBy(func(r *gen.GetJobStreamRequest) bool {
							// assert we're setting values in the request
							return r.GetJobId() == t.Name()
						}),
						([]grpc.CallOption)(nil),
					).
					Return(recvr, nil)
				return obj
			},
			writer: func(t *testing.T) *mockEventWriter {
				m := &mockEventWriter{}
				m.On("Write", mock.MatchedBy(func(e event.WaypointJobEvent) bool {
					return e.JobID == t.Name() && e.Type == event.WaypointJobEventTypeOpen
				})).Return(nil)
				return m
			},
			errExpected: fmt.Errorf("context canceled"),
		},

		"client error": {
			jobStreamGetter: func(t *testing.T, cnclFn func()) *mockJobStreamGetter {
				obj := &mockJobStreamGetter{}
				obj.
					On(
						"GetJobStream",
						mock.Anything,
						mock.MatchedBy(func(r *gen.GetJobStreamRequest) bool {
							// assert we're setting values in the request
							return r.GetJobId() == t.Name()
						}),
						([]grpc.CallOption)(nil),
					).
					Return(nil, fmt.Errorf("client error"))
				return obj
			},
			writer: func(t *testing.T) *mockEventWriter {
				m := &mockEventWriter{}
				m.On("Write", mock.MatchedBy(func(e event.WaypointJobEvent) bool {
					return e.JobID == t.Name() && e.Type == event.WaypointJobEventTypeOpen
				})).Return(nil)
				return m
			},
			errExpected: fmt.Errorf("client error"),
		},

		"error receiving": {
			jobStreamGetter: func(t *testing.T, cnclFn func()) *mockJobStreamGetter {
				recvr := &mockReceiver{}
				recvr.On("Recv").Return(nil, errConsumeWaypointDeploymentJobStream)

				obj := &mockJobStreamGetter{}
				obj.
					On(
						"GetJobStream",
						mock.Anything,
						mock.MatchedBy(func(r *gen.GetJobStreamRequest) bool {
							// assert we're setting values in the request
							return r.GetJobId() == t.Name()
						}),
						([]grpc.CallOption)(nil),
					).
					Return(recvr, nil)
				return obj
			},
			writer: func(t *testing.T) *mockEventWriter {
				m := &mockEventWriter{}
				m.On("Write", mock.MatchedBy(func(e event.WaypointJobEvent) bool {
					return e.JobID == t.Name() && e.Type == event.WaypointJobEventTypeOpen
				})).Return(nil)
				return m
			},
			errExpected: errConsumeWaypointDeploymentJobStream,
		},

		"stream error event": {
			jobStreamGetter: func(t *testing.T, cnclFn func()) *mockJobStreamGetter {
				recvr := &mockReceiver{}
				recvr.On("Recv").Return(&gen.GetJobStreamResponse{
					Event: &gen.GetJobStreamResponse_Error_{
						Error: &gen.GetJobStreamResponse_Error{
							Error: &status.Status{
								Code:    400,
								Message: "error",
							},
						},
					},
				}, nil)

				obj := &mockJobStreamGetter{}
				obj.
					On(
						"GetJobStream",
						mock.Anything,
						mock.MatchedBy(func(r *gen.GetJobStreamRequest) bool {
							// assert we're setting values in the request
							return r.GetJobId() == t.Name()
						}),
						([]grpc.CallOption)(nil),
					).
					Return(recvr, nil)
				return obj
			},
			writer: func(t *testing.T) *mockEventWriter {
				m := &mockEventWriter{}
				m.On("Write", mock.MatchedBy(func(e event.WaypointJobEvent) bool {
					return e.JobID == t.Name() && e.Type == event.WaypointJobEventTypeError
				})).Return(nil)
				return m
			},

			errExpected: event.ErrWaypointJobStream,
		},
		"error writing": {
			jobStreamGetter: func(t *testing.T, cnclFn func()) *mockJobStreamGetter {
				recvr := &mockReceiver{}
				recvr.On("Recv").Return(&gen.GetJobStreamResponse{
					Event: &gen.GetJobStreamResponse_Open_{
						Open: &gen.GetJobStreamResponse_Open{},
					},
				}, nil)

				obj := &mockJobStreamGetter{}
				obj.
					On(
						"GetJobStream",
						mock.Anything,
						mock.MatchedBy(func(r *gen.GetJobStreamRequest) bool {
							// assert we're setting values in the request
							return r.GetJobId() == t.Name()
						}),
						([]grpc.CallOption)(nil),
					).
					Return(recvr, nil)
				return obj
			},
			writer: func(t *testing.T) *mockEventWriter {
				m := &mockEventWriter{}
				m.On("Write", mock.MatchedBy(func(e event.WaypointJobEvent) bool {
					return e.JobID == t.Name() && e.Type == event.WaypointJobEventTypeOpen
				})).Return(errConsumeWaypointDeploymentJobStream)
				return m
			},
			errExpected: errConsumeWaypointDeploymentJobStream,
		},

		"job failure": {
			jobStreamGetter: func(t *testing.T, cnclFn func()) *mockJobStreamGetter {
				recvr := &mockReceiver{}
				recvr.On("Recv").Return(&gen.GetJobStreamResponse{
					Event: &gen.GetJobStreamResponse_Complete_{
						Complete: &gen.GetJobStreamResponse_Complete{
							Error: &status.Status{
								Message: "failure",
							},
							Result: &gen.Job_Result{},
						},
					},
				}, nil)

				obj := &mockJobStreamGetter{}
				obj.
					On(
						"GetJobStream",
						mock.Anything,
						mock.MatchedBy(func(r *gen.GetJobStreamRequest) bool {
							// assert we're setting values in the request
							return r.GetJobId() == t.Name()
						}),
						([]grpc.CallOption)(nil),
					).
					Return(recvr, nil)
				return obj
			},
			writer: func(t *testing.T) *mockEventWriter {
				m := &mockEventWriter{}
				m.On("Write", mock.MatchedBy(func(e event.WaypointJobEvent) bool {
					return e.JobID == t.Name() && e.Type == event.WaypointJobEventTypeComplete
				})).Return(nil)
				return m
			},
			errExpected: event.ErrWaypointJobFailed,
		},
	}

	for name, test := range tests {
		name := name
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, cancelFn := context.WithCancel(context.Background())

			g := test.jobStreamGetter(t, cancelFn)
			w := test.writer(t)
			p, err := New(v, WithClient(g), WithWriter(w), WithJobID(t.Name()))
			assert.NoError(t, err)

			err = p.Poll(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			g.AssertExpectations(t)
			w.AssertExpectations(t)
		})
	}
}
