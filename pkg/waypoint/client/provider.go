package client

import (
	"context"

	pb "github.com/hashicorp/waypoint/pkg/server/gen"
)

type Provider interface {
	Fetch(context.Context) (pb.WaypointClient, error)
	Close() error
}
