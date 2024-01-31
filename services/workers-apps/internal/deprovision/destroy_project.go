package deprovision

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/kube"
	waypoint "github.com/powertoolsdev/mono/pkg/waypoint/client"
	"github.com/powertoolsdev/mono/pkg/waypoint/client/k8s"
)

type DestroyWaypointProjectRequest struct {
	TokenSecretNamespace string           `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string           `json:"org_server_address" validate:"required"`
	OrgID                string           `json:"org_id" validate:"required"`
	AppID                string           `json:"install_id" validate:"required"`
	ClusterInfo          kube.ClusterInfo `json:"cluster_info" validate:"required"`
}

func (c DestroyWaypointProjectRequest) validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

type DestroyWaypointProjectResponse struct{}

func (a *Activities) DestroyWaypointProject(ctx context.Context, req DestroyWaypointProjectRequest) (DestroyWaypointProjectResponse, error) {
	var resp DestroyWaypointProjectResponse
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

	if err := a.destroyWaypointProject(ctx, client, req.AppID); err != nil {
		return resp, fmt.Errorf("unable to create waypoint project: %w", err)
	}

	return resp, nil
}

func (a *Activities) destroyWaypointProject(ctx context.Context, client gen.WaypointClient, installID string) error {
	req := &gen.DestroyProjectRequest{
		Project: &gen.Ref_Project{
			Project: installID,
		},
	}

	_, err := client.DestroyProject(ctx, req)
	return err
}
