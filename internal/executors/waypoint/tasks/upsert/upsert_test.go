package upsert

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type mockApplicationUpserter struct{ mock.Mock }

func (m *mockApplicationUpserter) UpsertApplication(
	ctx context.Context,
	in *gen.UpsertApplicationRequest,
	opts ...grpc.CallOption) (*gen.UpsertApplicationResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*gen.UpsertApplicationResponse), args.Error(1)
}

var _ applicationUpserter = (*mockApplicationUpserter)(nil)

func TestNew(t *testing.T) {
	t.Parallel()

	v := validator.New()

	tests := map[string]struct {
		v           *validator.Validate
		opts        []upserterOption
		errExpected error
	}{
		"happy path": {
			v: v,
			opts: []upserterOption{
				WithClient(&mockApplicationUpserter{}),
				WithProject("abc123"),
				WithName("component name"),
			},
		},
		"missing validator": {
			v: nil,
			opts: []upserterOption{
				WithClient(&mockApplicationUpserter{}),
				WithProject("abc123"),
				WithName("component name"),
			},
			errExpected: fmt.Errorf("validator is nil"),
		},
		"missing client": {
			v: v,
			opts: []upserterOption{
				WithProject("abc123"),
				WithName("component name"),
			},
			errExpected: fmt.Errorf("Field validation for 'Client' failed"),
		},
		"missing project": {
			v: v,
			opts: []upserterOption{
				WithClient(&mockApplicationUpserter{}),
				WithName("component name"),
			},
			errExpected: fmt.Errorf("Field validation for 'Project' failed"),
		},
		"missing name": {
			v: v,
			opts: []upserterOption{
				WithClient(&mockApplicationUpserter{}),
				WithProject("abc123"),
			},
			errExpected: fmt.Errorf("Field validation for 'Name' failed"),
		},
		"error on conifg": {
			v:           v,
			opts:        []upserterOption{func(*upserter) error { return fmt.Errorf("error on config") }},
			errExpected: fmt.Errorf("error on config"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			u, err := New(test.v, test.opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, "abc123", u.Project)
			assert.Equal(t, "component name", u.Name)
		})
	}
}

func TestUpserter_UpsertWaypointApplication(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		upserter    func(*testing.T) (*upserter, func(*testing.T))
		errExpected error
	}{
		"happy path": {
			upserter: func(t *testing.T) (*upserter, func(*testing.T)) {
				m := &mockApplicationUpserter{}
				m.
					On(
						"UpsertApplication",
						mock.Anything,
						mock.MatchedBy(func(r *gen.UpsertApplicationRequest) bool {
							return r.Name == t.Name() && r.GetProject().Project == "project"
						}),
						([]grpc.CallOption)(nil),
					).
					Return(&gen.UpsertApplicationResponse{}, nil)

				return &upserter{Client: m, Project: "project", Name: t.Name()}, func(t *testing.T) { m.AssertExpectations(t) }
			},
		},
		"on error": {
			upserter: func(t *testing.T) (*upserter, func(*testing.T)) {
				m := &mockApplicationUpserter{}
				m.
					On(
						"UpsertApplication",
						mock.Anything,
						mock.MatchedBy(func(r *gen.UpsertApplicationRequest) bool {
							return r.Name == t.Name() && r.GetProject().Project == "project"
						}),
						([]grpc.CallOption)(nil),
					).
					Return(&gen.UpsertApplicationResponse{}, fmt.Errorf("client error"))

				return &upserter{Client: m, Project: "project", Name: t.Name()}, func(t *testing.T) { m.AssertExpectations(t) }
			},
			errExpected: fmt.Errorf("client error"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			u, assertions := test.upserter(t)
			err := u.UpsertWaypointApplication(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assertions(t)
		})
	}
}
