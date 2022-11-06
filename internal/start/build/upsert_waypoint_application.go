package build

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"

	"github.com/powertoolsdev/go-waypoint"
)

type UpsertWaypointApplicationRequest struct {
	OrgID                string `json:"org_id" validate:"required"`
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`

	Component waypoint.Component `json:"component" validate:"required"`
}

func (u UpsertWaypointApplicationRequest) validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

type UpsertWaypointApplicationResponse struct{}

func (a *Activities) UpsertWaypointApplication(ctx context.Context, req UpsertWaypointApplicationRequest) (UpsertWaypointApplicationResponse, error) {
	var resp UpsertWaypointApplicationResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	client, err := a.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	if err := a.upsertWaypointApplication(ctx, client, req); err != nil {
		return resp, fmt.Errorf("failed to create waypoint application: %w", err)
	}

	return resp, nil
}

type waypointApplicationUpserter interface {
	upsertWaypointApplication(context.Context, waypointClientApplicationUpserter, UpsertWaypointApplicationRequest) error
}

type wpApplicationUpserter struct{}

var _ waypointApplicationUpserter = (*wpApplicationUpserter)(nil)

type waypointClientApplicationUpserter interface {
	UpsertApplication(ctx context.Context, in *gen.UpsertApplicationRequest, opts ...grpc.CallOption) (*gen.UpsertApplicationResponse, error)
}

func (*wpApplicationUpserter) upsertWaypointApplication(ctx context.Context, client waypointClientApplicationUpserter, req UpsertWaypointApplicationRequest) error {
	wpReq := &gen.UpsertApplicationRequest{
		Project: &gen.Ref_Project{
			Project: req.OrgID,
		},
		Name: req.Component.Name,
	}

	_, err := client.UpsertApplication(ctx, wpReq)
	if err != nil {
		return err
	}
	return nil
}
