package build

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
)

type UpsertWaypointWorkspaceRequest struct {
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`
	OrgID                string `json:"org_id" validate:"required"`
	AppID                string `json:"app_id" validate:"required"`
}

func (u UpsertWaypointWorkspaceRequest) validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

type UpsertWaypointWorkspaceResponse struct{}

func (a *Activities) UpsertWaypointWorkspace(ctx context.Context, req UpsertWaypointWorkspaceRequest) (UpsertWaypointWorkspaceResponse, error) {
	var resp UpsertWaypointWorkspaceResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	client, err := a.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	if err := a.upsertWaypointWorkspace(ctx, client, req.AppID, req.OrgID); err != nil {
		return resp, fmt.Errorf("failed to adopt waypoint runner: %w", err)
	}
	return resp, nil
}

type waypointWorkspaceUpserter interface {
	upsertWaypointWorkspace(context.Context, waypointClientWorkspaceUpserter, string, string) error
}

var _ waypointWorkspaceUpserter = (*wpWorkspaceUpserter)(nil)

type wpWorkspaceUpserter struct{}

type waypointClientWorkspaceUpserter interface {
	UpsertWorkspace(ctx context.Context, in *gen.UpsertWorkspaceRequest, opts ...grpc.CallOption) (*gen.UpsertWorkspaceResponse, error)
}

func (w *wpWorkspaceUpserter) upsertWaypointWorkspace(ctx context.Context, client waypointClientWorkspaceUpserter, appID, orgID string) error {
	req := &gen.UpsertWorkspaceRequest{
		Workspace: &gen.Workspace{
			Name: appID,
			Projects: []*gen.Workspace_Project{
				{
					Project: &gen.Ref_Project{
						Project: orgID,
					},
				},
			},
		},
	}

	_, err := client.UpsertWorkspace(ctx, req)
	return err
}
