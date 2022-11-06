package runner

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
)

const (
	defaultODRImagePullPolicy string = "Always"
	defaultODRImageURL        string = "hashicorp/waypoint-odr:0.10.2"
)

type CreateRunnerProfileRequest struct {
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`
	OrgID                string `json:"org_id" validate:"required"`
}

func (crp *CreateRunnerProfileRequest) validate() error {
	validate := validator.New()
	return validate.Struct(crp)
}

type CreateRunnerProfileResponse struct{}

func (a *Activities) CreateRunnerProfile(
	ctx context.Context,
	req CreateRunnerProfileRequest,
) (CreateRunnerProfileResponse, error) {
	var resp CreateRunnerProfileResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}
	client, err := a.waypointProvider.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	if err := createRunnerProfile(ctx, client, req); err != nil {
		return resp, fmt.Errorf("failed to create service config: %w", err)
	}

	return resp, nil
}

type clientODRConfigUpserter interface {
	UpsertOnDemandRunnerConfig(
		ctx context.Context,
		in *gen.UpsertOnDemandRunnerConfigRequest,
		opts ...grpc.CallOption,
	) (*gen.UpsertOnDemandRunnerConfigResponse, error)
}

func createRunnerProfile(ctx context.Context, client clientODRConfigUpserter, req CreateRunnerProfileRequest) error {
	imagePullPolicy := "Always"
	odrServiceAccount := fmt.Sprintf("waypoint-odr-%s", req.OrgID)

	_, err := client.UpsertOnDemandRunnerConfig(ctx, &gen.UpsertOnDemandRunnerConfigRequest{
		Config: &gen.OnDemandRunnerConfig{
			Name:       req.OrgID,
			OciUrl:     defaultODRImageURL,
			PluginType: "kubernetes",
			PluginConfig: []byte(fmt.Sprintf(`{
	"service_account": "%s",
	"image_pull_policy": "%s"
}`, odrServiceAccount, imagePullPolicy)),
			ConfigFormat:         gen.Hcl_JSON,
			Default:              true,
			EnvironmentVariables: map[string]string{},
		},
	})

	return err
}
