package runner

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/kube"
	waypoint "github.com/powertoolsdev/mono/pkg/waypoint/client"
	"github.com/powertoolsdev/mono/pkg/waypoint/client/k8s"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GetWaypointServerCookieRequest struct {
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`
	OrgID                string `json:"org_id" validate:"required"`

	ClusterInfo kube.ClusterInfo `json:"cluster_info" validate:"required"`
}

func (g GetWaypointServerCookieRequest) validate() error {
	validate := validator.New()
	return validate.Struct(g)
}

type GetWaypointServerCookieResponse struct {
	Cookie string `json:"cookie"`
}

func (a *Activities) GetWaypointServerCookie(ctx context.Context, req GetWaypointServerCookieRequest) (GetWaypointServerCookieResponse, error) {
	var resp GetWaypointServerCookieResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	provider, err := k8s.New(a.v, k8s.WithConfig(k8s.Config{
		Address: req.OrgServerAddr,
		Token: k8s.Token{
			Namespace: req.TokenSecretNamespace,
			Name:      waypoint.DefaultTokenSecretName(req.OrgID),
			Key:       waypoint.DefaultTokenSecretKey,
		},
		ClusterInfo: &req.ClusterInfo,
	}))
	if err != nil {
		return resp, fmt.Errorf("unable to get org provider: %w", err)
	}

	client, err := provider.Fetch(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to get client: %w", err)
	}

	cookie, err := a.getWaypointServerCookie(ctx, client)
	if err != nil {
		return resp, fmt.Errorf("unable to get waypoint runner cookie: %w", err)
	}
	resp.Cookie = cookie
	return resp, nil
}

type waypointServerCookieGetter interface {
	getWaypointServerCookie(context.Context, waypointClientServerConfigGetter) (string, error)
}

var _ waypointServerCookieGetter = (*wpServerCookieGetter)(nil)

type wpServerCookieGetter struct{}

type waypointClientServerConfigGetter interface {
	GetServerConfig(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*gen.GetServerConfigResponse, error)
}

func (w *wpServerCookieGetter) getWaypointServerCookie(ctx context.Context, client waypointClientServerConfigGetter) (string, error) {
	resp, err := client.GetServerConfig(ctx, &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return resp.Config.Cookie, nil
}
