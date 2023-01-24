package upsert

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
)

type applicationUpserter interface {
	UpsertApplication(ctx context.Context, in *gen.UpsertApplicationRequest, opts ...grpc.CallOption) (*gen.UpsertApplicationResponse, error)
}

var _ applicationUpserter = (gen.WaypointClient)(nil)

type upserter struct {
	Client  applicationUpserter `validate:"required"`
	Project string              `validate:"required"`
	Name    string              `validate:"required"`

	// internal state
	v *validator.Validate
}

type upserterOption func(*upserter) error

func New(v *validator.Validate, opts ...upserterOption) (*upserter, error) {
	u := &upserter{v: v}

	if v == nil {
		return nil, fmt.Errorf("error instantiating executor: validator is nil")
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

func WithClient(c applicationUpserter) upserterOption {
	return func(u *upserter) error {
		u.Client = c
		return nil
	}
}

func WithProject(id string) upserterOption {
	return func(u *upserter) error {
		u.Project = id
		return nil
	}
}

func WithName(name string) upserterOption {
	return func(u *upserter) error {
		u.Name = name
		return nil
	}
}

func (u *upserter) UpsertWaypointApplication(ctx context.Context) error {
	req := &gen.UpsertApplicationRequest{
		Project: &gen.Ref_Project{
			Project: u.Project,
		},
		Name: u.Name,
	}

	_, err := u.Client.UpsertApplication(ctx, req)
	return err
}
