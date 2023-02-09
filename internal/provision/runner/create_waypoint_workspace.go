package runner

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
)

type CreateWaypointWorkspaceRequest struct {
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`
	OrgID                string `json:"org_id" validate:"required"`
	InstallID            string `json:"install_id" validate:"required"`
}

func (c CreateWaypointWorkspaceRequest) validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

type CreateWaypointWorkspaceResponse struct{}

func (a *Activities) CreateWaypointWorkspace(ctx context.Context, req CreateWaypointWorkspaceRequest) (CreateWaypointWorkspaceResponse, error) {
	var resp CreateWaypointWorkspaceResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	client, err := a.waypointProvider.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	if err := a.createWaypointWorkspace(ctx, client, req.InstallID); err != nil {
		return resp, fmt.Errorf("failed to adopt waypoint runner: %w", err)
	}
	return resp, nil
}

type waypointWorkspaceCreator interface {
	createWaypointWorkspace(context.Context, waypointClientWorkspaceUpserter, string) error
}

var _ waypointWorkspaceCreator = (*wpWorkspaceCreator)(nil)

type wpWorkspaceCreator struct{}

type waypointClientWorkspaceUpserter interface {
	UpsertWorkspace(ctx context.Context, in *gen.UpsertWorkspaceRequest, opts ...grpc.CallOption) (*gen.UpsertWorkspaceResponse, error)
}

func (w *wpWorkspaceCreator) createWaypointWorkspace(ctx context.Context, client waypointClientWorkspaceUpserter, installID string) error {
	req := &gen.UpsertWorkspaceRequest{
		Workspace: &gen.Workspace{
			Name: installID,
			Projects: []*gen.Workspace_Project{
				{
					Project: &gen.Ref_Project{
						Project: installID,
					},
				},
			},
		},
	}

	_, err := client.UpsertWorkspace(ctx, req)
	return err
}
