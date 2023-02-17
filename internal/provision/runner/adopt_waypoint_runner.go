package runner

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AdoptWaypointRunnerRequest struct {
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`
	OrgID                string `json:"org_id" validate:"required"`
	InstallID            string `json:"install_id" validate:"required"`
}

func (a AdoptWaypointRunnerRequest) validate() error {
	validate := validator.New()
	return validate.Struct(a)
}

type AdoptWaypointRunnerResponse struct{}

func (a *Activities) AdoptWaypointRunner(ctx context.Context, req AdoptWaypointRunnerRequest) (AdoptWaypointRunnerResponse, error) {
	var resp AdoptWaypointRunnerResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	client, err := a.Provider.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	if err := a.adoptWaypointRunner(ctx, client, req.InstallID); err != nil {
		return resp, fmt.Errorf("failed to adopt waypoint runner: %w", err)
	}
	return resp, nil
}

type waypointRunnerAdopter interface {
	adoptWaypointRunner(context.Context, waypointClientRunnerAdopter, string) error
}

var _ waypointRunnerAdopter = (*wpRunnerAdopter)(nil)

type wpRunnerAdopter struct{}

type waypointClientRunnerAdopter interface {
	AdoptRunner(ctx context.Context, in *gen.AdoptRunnerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

func (w *wpRunnerAdopter) adoptWaypointRunner(ctx context.Context, client waypointClientRunnerAdopter, installID string) error {
	req := &gen.AdoptRunnerRequest{
		RunnerId: installID,
		Adopt:    true,
	}

	_, err := client.AdoptRunner(ctx, req)
	return err
}
