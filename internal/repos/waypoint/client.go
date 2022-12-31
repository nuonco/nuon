package waypoint

import (
	"context"
	"fmt"

	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=client_mocks_test.go -source=client.go -package=waypoint
type waypointClient interface {
	waypointv1.WaypointClient
}

// waypointClientProvider is the wrapper we use to talk to make waypoint clients using kubernetes
type waypointClientProvider interface {
	GetOrgWaypointClient(context.Context, string, string, string) (waypointv1.WaypointClient, error)
}

// clientGetter is the local function we use to stub out this method from other parts of the repo
type clientGetter = func(context.Context) (waypointClient, error)

func (r *repo) getClient(ctx context.Context) (waypointClient, error) {
	orgCtx, err := r.CtxGetter(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get context: %w", err)
	}

	// TODO(jm): update the `go-waypoint` client to accept both a kube cluster info and the secret name instead of
	// being "magical".
	client, err := r.WaypointClientProvider.GetOrgWaypointClient(ctx,
		orgCtx.WaypointServer.SecretNamespace,
		orgCtx.OrgID,
		orgCtx.WaypointServer.Address)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return client, nil
}
