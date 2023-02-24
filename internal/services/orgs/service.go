package services

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/orgs-api/internal/orgcontext"
	"github.com/powertoolsdev/orgs-api/internal/repos/waypoint"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=service_mocks.go -source=service.go -package=services
type Service interface {
	GetInfo(context.Context, string) (*orgsv1.GetInfoResponse, error)
	GetRunners(context.Context, string) (*orgsv1.GetRunnersResponse, error)
}

func New(opts ...serviceOption) (*service, error) {
	srv := &service{
		ctxGetter: orgcontext.Get,
	}
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
	WaypointRepo  waypoint.Repo  `validate:"required"`
	WorkflowsRepo workflows.Repo `validate:"required"`

	// this is only settable for internal testing purposes
	ctxGetter func(context.Context) (*orgcontext.Context, error)
}

var _ Service = (*service)(nil)

type serviceOption func(*service) error

func WithWorkflowsRepo(repo workflows.Repo) serviceOption {
	return func(s *service) error {
		s.WorkflowsRepo = repo
		return nil
	}
}

func WithWaypointRepo(repo waypoint.Repo) serviceOption {
	return func(s *service) error {
		s.WaypointRepo = repo
		return nil
	}
}
