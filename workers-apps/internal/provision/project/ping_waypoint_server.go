package project

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/go-waypoint"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PingWaypointServerRequest struct {
	Timeout time.Duration `json:"timeout" validate:"required,gt=0m"`
	Addr    string        `json:"addr" validate:"required"`
}

func (p *PingWaypointServerRequest) validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type PingWaypointServerResponse struct{}

// PingWaypointServer pings the waypoint server until it's responding, up to the timeout
func (a *Activities) PingWaypointServer(ctx context.Context, req PingWaypointServerRequest) (PingWaypointServerResponse, error) {
	var resp PingWaypointServerResponse

	ctx, cancelFn := context.WithTimeout(ctx, req.Timeout)
	defer cancelFn()

	err := a.pingWaypointServerUntilReachable(ctx, req.Addr, a.Provider)
	if err != nil {
		return resp, fmt.Errorf("unable to reach waypoint server after %s: %w", req.Timeout, err)
	}

	return resp, nil
}

type waypointServerPinger interface {
	pingWaypointServerUntilReachable(context.Context, string, waypoint.Provider) error
}

type wpServerPinger struct{}

var _ waypointServerPinger = (*wpServerPinger)(nil)

var errUnableToPingWaypointServer = fmt.Errorf("unable to ping waypoint server")

// pingUntilReachable pings a server until its reachable
func (w *wpServerPinger) pingWaypointServerUntilReachable(ctx context.Context, addr string, provider waypoint.Provider) error {
	for {
		select {
		case <-ctx.Done():
			return errUnableToPingWaypointServer
		default:
		}

		client, err := provider.GetUnauthenticatedWaypointClient(ctx, addr)
		if err != nil {
			fmt.Printf("unable to get client: %v", err)
			continue
		}

		_, err = w.getVersionInfo(ctx, client)
		if err == nil {
			break
		}
		fmt.Printf("unable to get version: %v", err)
	}
	return nil
}

func (w *wpServerPinger) getVersionInfo(ctx context.Context, client waypointClientGetVersionInfo) (*gen.GetVersionInfoResponse, error) {
	return client.GetVersionInfo(ctx, &emptypb.Empty{})
}

type waypointClientGetVersionInfo interface {
	GetVersionInfo(context.Context, *emptypb.Empty, ...grpc.CallOption) (*gen.GetVersionInfoResponse, error)
}
