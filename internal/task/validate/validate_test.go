package validate

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

type mockJobValidator struct{ mock.Mock }

func (m *mockJobValidator) ValidateJob(
	ctx context.Context,
	in *gen.ValidateJobRequest,
	opts ...grpc.CallOption) (*gen.ValidateJobResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*gen.ValidateJobResponse), args.Error(1)
}

var _ jobValidator = (*mockJobValidator)(nil)

func TestNew(t *testing.T) {
	t.Parallel()

	v := validator.New()

	tests := map[string]struct {
		v           *validator.Validate
		opts        []validaterOption
		errExpected error
	}{
		"happy path": {
			v: v,
			opts: []validaterOption{
				WithClient(&mockJobValidator{}),
				WithID("abc123"),
			},
		},
		"missing validator": {
			v: nil,
			opts: []validaterOption{
				WithClient(&mockJobValidator{}),
				WithID("abc123"),
			},
			errExpected: fmt.Errorf("validator is nil"),
		},
		"missing client": {
			v: v,
			opts: []validaterOption{
				WithID("abc123"),
			},
			errExpected: fmt.Errorf("Field validation for 'Client' failed"),
		},
		"missing id": {
			v: v,
			opts: []validaterOption{
				WithClient(&mockJobValidator{}),
			},
			errExpected: fmt.Errorf("Field validation for 'ID' failed"),
		},
		"error on conifg": {
			v:           v,
			opts:        []validaterOption{func(*validater) error { return fmt.Errorf("error on config") }},
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
			assert.Equal(t, "abc123", u.ID)
		})
	}
}

func TestValidater_Validate(t *testing.T) {
	t.Parallel()

	v := validator.New()

	tests := map[string]struct {
		validater   func(*testing.T) *mockJobValidator
		errExpected error
	}{
		"happy path": {
			validater: func(t *testing.T) *mockJobValidator {
				m := &mockJobValidator{}
				m.
					On(
						"ValidateJob",
						mock.Anything,
						mock.MatchedBy(func(r *gen.ValidateJobRequest) bool {
							// assert we're setting values in the request
							return r.GetJob().Id == t.Name()
						}),
						([]grpc.CallOption)(nil),
					).
					Return(&gen.ValidateJobResponse{}, nil)

				return m
			},
		},
		"on error": {
			validater: func(t *testing.T) *mockJobValidator {
				m := &mockJobValidator{}
				m.
					On(
						"ValidateJob",
						mock.Anything,
						mock.MatchedBy(func(r *gen.ValidateJobRequest) bool {
							// assert we're setting values in the request
							return r.GetJob().Id == t.Name()
						}),
						([]grpc.CallOption)(nil),
					).
					Return(&gen.ValidateJobResponse{}, fmt.Errorf("client error"))

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

			m := test.validater(t)
			v, err := New(v, WithClient(m), WithID(t.Name()))
			assert.NoError(t, err)

			err = v.Validate(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			m.AssertExpectations(t)
		})
	}
}
