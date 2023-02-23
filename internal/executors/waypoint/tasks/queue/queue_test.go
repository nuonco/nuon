package queue

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type mockJobQueuer struct{ mock.Mock }

func (m *mockJobQueuer) QueueJob(
	ctx context.Context,
	in *gen.QueueJobRequest,
	opts ...grpc.CallOption) (*gen.QueueJobResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*gen.QueueJobResponse), args.Error(1)
}

func (m *mockJobQueuer) GetLatestPushedArtifact(ctx context.Context, in *gen.GetLatestPushedArtifactRequest, opts ...grpc.CallOption) (*gen.PushedArtifact, error) {
	return nil, nil
}

var _ jobQueuer = (*mockJobQueuer)(nil)

func TestNew(t *testing.T) {
	t.Parallel()

	v := validator.New()

	tests := map[string]struct {
		v              *validator.Validate
		additionalOpts []queuerOption
		errExpected    error
	}{
		"happy path": {
			v: v,
			additionalOpts: []queuerOption{
				WithClient(&mockJobQueuer{}),
				WithProject("abc123"),
			},
		},
		"missing validator": {
			v:              nil,
			additionalOpts: []queuerOption{},
			errExpected:    fmt.Errorf("validator is nil"),
		},
		"missing client": {
			v:              v,
			additionalOpts: []queuerOption{WithClient(nil)},
			errExpected:    fmt.Errorf("Field validation for 'Client' failed"),
		},
		"missing id": {
			v:              v,
			additionalOpts: []queuerOption{WithID("")},
			errExpected:    fmt.Errorf("Field validation for 'ID' failed"),
		},
		"error on config": {
			v:              v,
			additionalOpts: []queuerOption{func(*queuer) error { return fmt.Errorf("error on config") }},
			errExpected:    fmt.Errorf("error on config"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := append(testOptions(), test.additionalOpts...)

			q, err := New(test.v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, q)
		})
	}
}

func testOptions() []queuerOption {
	return []queuerOption{
		WithClient(&mockJobQueuer{}),
		WithID("id"),
		WithWorkspace("workspace"),
		WithProject("project"),
		WithApplication("application-id"),
		WithLabels(map[string]string{"test": "label"}),
		WithGitURL("https://github.com/powertoolsdev/jmorehouse/empty"),
		WithTargetRunnerID("target-runner-id"),
		WithOnDemandRunnerName("on-demand-runner-name"),
		WithJobTimeout("1m"),
		WithJobType(planv1.WaypointJobType_WAYPOINT_JOB_TYPE_BUILD),
		WithWaypointHCL([]byte("waypoint-hcl")),
	}
}

func TestUpserter_UpsertWaypointApplication(t *testing.T) {
	t.Parallel()

	v := validator.New()

	tests := map[string]struct {
		queuer      func(*testing.T) *mockJobQueuer
		errExpected error
	}{
		"happy path": {
			queuer: func(t *testing.T) *mockJobQueuer {
				m := &mockJobQueuer{}
				m.
					On(
						"QueueJob",
						mock.Anything,
						mock.MatchedBy(func(r *gen.QueueJobRequest) bool {
							j := r.GetJob()

							// assert we're setting values in the job request
							return r.ExpiresIn == "1m" && j.SingletonId == "id"
						}),
						([]grpc.CallOption)(nil),
					).
					Return(&gen.QueueJobResponse{JobId: t.Name()}, nil)

				return m
			},
		},
		"on error": {
			queuer: func(t *testing.T) *mockJobQueuer {
				m := &mockJobQueuer{}
				m.
					On(
						"QueueJob",
						mock.Anything,
						mock.MatchedBy(func(r *gen.QueueJobRequest) bool {
							j := r.GetJob()

							// assert we're setting values in the job request
							return r.ExpiresIn == "1m" && j.SingletonId == "id"
						}),
						([]grpc.CallOption)(nil),
					).
					Return(&gen.QueueJobResponse{JobId: t.Name()}, fmt.Errorf("client error"))

				return m
			},
			errExpected: fmt.Errorf("client error"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := test.queuer(t)
			opts := append(testOptions(), WithClient(m))

			q, err := New(v, opts...)
			assert.NoError(t, err)

			id, err := q.QueueDeployment(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, id, t.Name())
			m.AssertExpectations(t)
		})
	}
}
