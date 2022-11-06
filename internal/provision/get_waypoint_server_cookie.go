package provision

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GetWaypointServerCookieRequest struct {
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`
	OrgID                string `json:"org_id" validate:"required"`
}

func validateGetWaypointServerCookieRequest(req GetWaypointServerCookieRequest) error {
	validate := validator.New()
	return validate.Struct(req)
}

type GetWaypointServerCookieResponse struct {
	Cookie string `json:"cookie"`
}

func (a *ProvisionActivities) GetWaypointServerCookie(ctx context.Context, req GetWaypointServerCookieRequest) (GetWaypointServerCookieResponse, error) {
	var resp GetWaypointServerCookieResponse
	if err := validateGetWaypointServerCookieRequest(req); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	client, err := a.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	cookie, err := a.getWaypointServerCookie(ctx, client)
	if err != nil {
		return resp, fmt.Errorf("unable to create waypoint runner cookie: %w", err)
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
