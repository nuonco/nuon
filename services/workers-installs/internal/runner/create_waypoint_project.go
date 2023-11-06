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
)

type CreateWaypointProjectRequest struct {
	TokenSecretNamespace string           `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string           `json:"org_server_address" validate:"required"`
	OrgID                string           `json:"org_id" validate:"required"`
	InstallID            string           `json:"install_id" validate:"required"`
	ClusterInfo          kube.ClusterInfo `json:"cluster_info" validate:"required"`
}

func (c CreateWaypointProjectRequest) validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

type CreateWaypointProjectResponse struct{}

func (a *Activities) CreateWaypointProject(ctx context.Context, req CreateWaypointProjectRequest) (CreateWaypointProjectResponse, error) {
	var resp CreateWaypointProjectResponse
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

	if err := a.createWaypointProject(ctx, client, req.InstallID); err != nil {
		return resp, fmt.Errorf("unable to create waypoint project: %w", err)
	}

	return resp, nil
}

type waypointProjectCreator interface {
	createWaypointProject(context.Context, waypointClientProjectUpserter, string) error
}

var _ waypointProjectCreator = (*wpProjectCreator)(nil)

type wpProjectCreator struct{}

type waypointClientProjectUpserter interface {
	UpsertProject(ctx context.Context, in *gen.UpsertProjectRequest, opts ...grpc.CallOption) (*gen.UpsertProjectResponse, error)
}

func (w *wpProjectCreator) createWaypointProject(ctx context.Context, client waypointClientProjectUpserter, installID string) error {
	waypointHcl, err := getProjectWaypointConfig(installID)
	if err != nil {
		return fmt.Errorf("unable to create project waypoint config: %w", err)
	}

	req := &gen.UpsertProjectRequest{
		Project: &gen.Project{
			Name:          installID,
			RemoteEnabled: true,
			DataSource: &gen.Job_DataSource{
				Source: &gen.Job_DataSource_Git{
					// NOTE(jm): this is a temporary hack until we either a.) figure out a way to
					// not pass a repo in, or b.) figure out how to have different data sources in
					// waypoint
					Git: &gen.Job_Git{
						Url: "https://github.com/jonmorehouse/empty",
					},
				},
			},
			DataSourcePoll: &gen.Project_Poll{
				Enabled: false,
			},
			WaypointHcl:       waypointHcl,
			WaypointHclFormat: gen.Hcl_JSON,
		},
	}

	_, err = client.UpsertProject(ctx, req)
	return err
}
