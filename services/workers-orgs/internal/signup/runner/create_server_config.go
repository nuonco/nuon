package runner

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
)

type CreateServerConfigRequest struct {
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`
	OrgID                string `json:"org_id" validate:"required"`
}

type CreateServerConfigResponse struct{}

func (r *CreateServerConfigRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

func (a *Activities) CreateServerConfig(
	ctx context.Context,
	req CreateServerConfigRequest,
) (CreateServerConfigResponse, error) {
	var resp CreateServerConfigResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	client, err := a.waypointProvider.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	if err := createServerConfig(ctx, client, req); err != nil {
		return resp, fmt.Errorf("failed to create service config: %w", err)
	}

	return resp, nil
}

func createServerConfig(ctx context.Context, client gen.WaypointClient, req CreateServerConfigRequest) error {
	_, err := client.SetServerConfig(ctx, &gen.SetServerConfigRequest{
		Config: &gen.ServerConfig{
			AdvertiseAddrs: []*gen.ServerConfig_AdvertiseAddr{
				{
					Addr:          req.OrgServerAddr,
					Tls:           true,
					TlsSkipVerify: true,
				},
			},
			Platform: "kubernetes",
		},
	})
	return err
}
