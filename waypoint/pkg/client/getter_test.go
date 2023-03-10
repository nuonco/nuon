package client

import (
	"context"

	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type mockClientGetter struct{ mock.Mock }

func (m *mockClientGetter) getClient(ctx context.Context, addr, token string) (*grpc.ClientConn, pb.WaypointClient, error) {
	args := m.Called(ctx, addr, token)
	return args.Get(0).(*grpc.ClientConn), args.Get(1).(pb.WaypointClient), args.Error(2)
}

var _ clientGetter = (*mockClientGetter)(nil)
