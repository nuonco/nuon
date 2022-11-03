package waypoint

import (
	"context"

	"github.com/hashicorp/go-hclog"
	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/hashicorp/waypoint/pkg/serverclient"
	"google.golang.org/grpc"
)

const (
	defaultHCLogName string = "nuon"
)

type Provider interface {
	// GetUnauthenticatedWaypointClient returns an unauthenticated waypoint client for making calls like get
	// version, or bootstrap.
	GetUnauthenticatedWaypointClient(context.Context, string) (pb.WaypointClient, error)

	// GetOrgWaypointClient returns a client with a token to talk to the specified org server. It loads the token
	// from the expected secret, in the passed in namespace
	GetOrgWaypointClient(context.Context, string, string, string) (pb.WaypointClient, error)
}

func NewProvider() Provider {
	return &wpClientProvider{
		connector:   &wpServerConnector{},
		tokenGetter: &k8sTokenGetter{},
	}
}

type wpClientProvider struct {
	connector serverConnector
	tokenGetter
}

var _ Provider = (*wpClientProvider)(nil)

// GetUnauthenticatedWaypointClient returns a waypoint client with no configured token
func (w *wpClientProvider) GetUnauthenticatedWaypointClient(ctx context.Context, addr string) (pb.WaypointClient, error) {
	return w.getClient(ctx, addr, "")
}

func (w *wpClientProvider) GetOrgWaypointClient(
	ctx context.Context,
	secretNamespace, orgID, addr string,
) (pb.WaypointClient, error) {
	token, err := w.getOrgToken(ctx, secretNamespace, orgID)
	if err != nil {
		return nil, err
	}

	return w.getClient(ctx, addr, token)
}

// getClient returns a waypoint client
func (w *wpClientProvider) getClient(ctx context.Context, addr, token string) (pb.WaypointClient, error) {
	cfg, err := serverclient.ContextConfig()
	if err != nil {
		return nil, err
	}
	cfg.Server.RequireAuth = false
	cfg.Server.Address = addr
	cfg.Server.Tls = true
	cfg.Server.TlsSkipVerify = true
	cfg.Server.RequireAuth = token != ""
	cfg.Server.AuthToken = token
	primaryOpts := serverclient.FromContextConfig(cfg)

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  defaultHCLogName,
		Level: hclog.LevelFromString("DEBUG"),
	})
	logOpt := serverclient.Logger(appLogger)

	cc, err := w.connector.Connect(ctx, primaryOpts, logOpt)
	if err != nil {
		return nil, err
	}

	client := pb.NewWaypointClient(cc)
	return client, nil
}

// tokenGetter provides a way to get tokens for a specific org
type tokenGetter interface {
	getOrgToken(context.Context, string, string) (string, error)
}

// server connector is the interface that we use to connect to a waypoint server
type serverConnector interface {
	Connect(context.Context, ...serverclient.ConnectOption) (*grpc.ClientConn, error)
}

var _ serverConnector = (*wpServerConnector)(nil)

// the wpServerConnector is a light wrapper around the package level connect function
type wpServerConnector struct{}

func (w *wpServerConnector) Connect(ctx context.Context, opts ...serverclient.ConnectOption) (*grpc.ClientConn, error) {
	return serverclient.Connect(ctx, opts...)
}
