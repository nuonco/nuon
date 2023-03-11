package client

import (
	"context"

	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
)

type clientGetter interface {
	getClient(context.Context, string, string) (*grpc.ClientConn, pb.WaypointClient, error)
}
type defaultClientGetter struct{}

func (g *defaultClientGetter) getClient(ctx context.Context, addr, token string) (*grpc.ClientConn, pb.WaypointClient, error) {
	cc, err := getClient(ctx, addr, token)
	if err != nil {
		return nil, nil, err
	}

	client := pb.NewWaypointClient(cc)
	return cc, client, nil
}

var _ clientGetter = (*defaultClientGetter)(nil)
