package waypoint

import (
	"context"

	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
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

type repo struct{}

var _ Repo = (*repo)(nil)
