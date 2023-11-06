package project

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/kube"
	waypoint "github.com/powertoolsdev/mono/pkg/waypoint/client"
	"github.com/powertoolsdev/mono/pkg/waypoint/client/k8s"
	"google.golang.org/grpc"
)

type UpsertWaypointWorkspaceRequest struct {
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`
	OrgID                string `json:"org_id" validate:"required"`
	AppID                string `json:"app_id" validate:"required"`

	ClusterInfo kube.ClusterInfo `json:"cluster_info" validate:"required"`
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

	if err := a.upsertWaypointWorkspace(ctx, client, req.AppID); err != nil {
		return resp, fmt.Errorf("failed to adopt waypoint runner: %w", err)
	}
	return resp, nil
}

type waypointWorkspaceUpserter interface {
	upsertWaypointWorkspace(context.Context, waypointClientWorkspaceUpserter, string) error
}

var _ waypointWorkspaceUpserter = (*wpWorkspaceUpserter)(nil)

type wpWorkspaceUpserter struct{}

type waypointClientWorkspaceUpserter interface {
	UpsertWorkspace(ctx context.Context, in *gen.UpsertWorkspaceRequest, opts ...grpc.CallOption) (*gen.UpsertWorkspaceResponse, error)
}

func (w *wpWorkspaceUpserter) upsertWaypointWorkspace(ctx context.Context, client waypointClientWorkspaceUpserter, appID string) error {
	req := &gen.UpsertWorkspaceRequest{
		Workspace: &gen.Workspace{
			Name: appID,
			Projects: []*gen.Workspace_Project{
				{
					Project: &gen.Ref_Project{
						Project: appID,
					},
				},
			},
		},
	}

	_, err := client.UpsertWorkspace(ctx, req)
	return err
}
