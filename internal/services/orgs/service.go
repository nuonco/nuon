package services

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/orgs-api/internal/repos/s3"
	"github.com/powertoolsdev/orgs-api/internal/repos/waypoint"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=service_mocks.go -source=service.go -package=services
type Service interface {
	GetInfo(context.Context, string) (*orgsv1.GetInfoResponse, error)
}

func New(opts ...serviceOption) (*service, error) {
	srv := &service{}
	for _, opt := range opts {
		if err := opt(srv); err != nil {
			return nil, err
		}
	}

	v := validator.New()
	if err := v.Struct(srv); err != nil {
		return nil, err
	}

	return srv, nil
}

type service struct {
	S3Repo       s3.Repo       `validate:"required"`
	WaypointRepo waypoint.Repo `validate:"required"`
}

var _ Service = (*service)(nil)

type serviceOption func(*service) error

func WithS3Repo(repo s3.Repo) serviceOption {
	return func(s *service) error {
		s.S3Repo = repo
		return nil
	}
}

func WithWaypointRepo(repo waypoint.Repo) serviceOption {
	return func(s *service) error {
		s.WaypointRepo = repo
		return nil
	}
}
