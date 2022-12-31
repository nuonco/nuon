package waypoint

import (
	"context"
	"fmt"

	"github.com/go-playground/validator"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/go-waypoint"
	"github.com/powertoolsdev/orgs-api/internal/orgcontext"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=waypoint_mocks.go -source=waypoint.go -package=waypoint
type Repo interface {
	GetVersionInfo(context.Context) (*waypointv1.GetVersionInfoResponse, error)
	GetRunner(context.Context, string) (*waypointv1.Runner, error)
	ListRunners(context.Context) (*waypointv1.ListRunnersResponse, error)
	ListJobs(context.Context) (*waypointv1.ListJobsResponse, error)
}

type orgContextGetter = func(context.Context) (*orgcontext.Context, error)

type repo struct {
	// the following fields are set for testing purposes
	CtxGetter              orgContextGetter       `validate:"required"`
	ClientGetter           clientGetter           `validate:"required"`
	WaypointClientProvider waypointClientProvider `validate:"required"`
}

var _ Repo = (*repo)(nil)

func New() (*repo, error) {
	v := validator.New()
	r := &repo{
		CtxGetter:              orgcontext.Get,
		WaypointClientProvider: waypoint.NewProvider(),
	}
	r.ClientGetter = r.getClient

	if err := v.Struct(r); err != nil {
		return nil, fmt.Errorf("unable to validate repo: %w", err)
	}

	return r, nil
}
