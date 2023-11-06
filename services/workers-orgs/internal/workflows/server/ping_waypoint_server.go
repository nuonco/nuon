package server

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/waypoint/client/public"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PingWaypointServerRequest struct {
	Timeout time.Duration `json:"timeout" validate:"required,gt=0m"`
	Addr    string        `json:"addr" validate:"required"`
}

func validatePingWaypointServerRequest(req PingWaypointServerRequest) error {
	validate := validator.New()
	return validate.Struct(req)
}

type PingWaypointServerResponse struct{}

// PingWaypointServer pings the waypoint server until it's responding, up to the timeout
func (a *Activities) PingWaypointServer(ctx context.Context, req PingWaypointServerRequest) (PingWaypointServerResponse, error) {
	var resp PingWaypointServerResponse

	ctx, cancelFn := context.WithTimeout(ctx, req.Timeout)
	defer cancelFn()

	provider, err := public.New(a.v, public.WithAddress(req.Addr))
	if err != nil {
		return resp, fmt.Errorf("unable to get org provider: %w", err)
	}

	err = a.pingWaypointServerUntilReachable(ctx, provider)
	if err != nil {
		return resp, fmt.Errorf("unable to reach waypoint server after %s: %w", req.Timeout, err)
	}

	return resp, nil
}

type waypointProvider interface {
	Fetch(context.Context) (gen.WaypointClient, error)
}

type waypointServerPinger interface {
	pingWaypointServerUntilReachable(context.Context, waypointProvider) error
}

type wpServerPinger struct{}

var _ waypointServerPinger = (*wpServerPinger)(nil)

var errUnableToPingWaypointServer = fmt.Errorf("unable to ping waypoint server")

// pingUntilReachable pings a server until its reachable
func (w *wpServerPinger) pingWaypointServerUntilReachable(ctx context.Context, provider waypointProvider) error {
	for {
		select {
		case <-ctx.Done():
			return errUnableToPingWaypointServer
		default:
		}

		client, err := provider.Fetch(ctx)
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
