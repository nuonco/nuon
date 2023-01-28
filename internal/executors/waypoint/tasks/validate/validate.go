package validate

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
)

type jobValidator interface {
	ValidateJob(ctx context.Context, in *gen.ValidateJobRequest, opts ...grpc.CallOption) (*gen.ValidateJobResponse, error)
}

var _ jobValidator = (gen.WaypointClient)(nil)

type validater struct {
	Client jobValidator `validate:"required"`
	ID     string       `validate:"required"`

	// internal state
	v *validator.Validate
}

type validaterOption func(*validater) error

func New(v *validator.Validate, opts ...validaterOption) (*validater, error) {
	u := &validater{v: v}

	if v == nil {
		return nil, fmt.Errorf("error instantiating validate task: validator is nil")
	}

	for _, opt := range opts {
		if err := opt(u); err != nil {
			return nil, err
		}
	}

	if err := u.v.Struct(u); err != nil {
		return nil, err
	}

	return u, nil
}

func WithClient(c jobValidator) validaterOption {
	return func(u *validater) error {
		u.Client = c
		return nil
	}
}

func WithID(id string) validaterOption {
	return func(u *validater) error {
		u.ID = id
		return nil
	}
}

func (v *validater) Validate(ctx context.Context) error {
	_, err := v.Client.ValidateJob(ctx, &gen.ValidateJobRequest{
		Job: &gen.Job{
			Id: v.ID,
		},
	})
	return err
}
