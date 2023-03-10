package waypoint

import (
	"context"

	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=client_mocks_test.go -source=client.go -package=waypoint
type waypointClient interface {
	waypointv1.WaypointClient
}

type waypointClientProvider interface {
	GetClient(context.Context) (waypointv1.WaypointClient, error)
}
