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

type CreateWaypointRunnerProfileRequest struct {
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`
	OrgID                string `json:"org_id" validate:"required"`
	InstallID            string `json:"install_id" validate:"required"`
}

func (c CreateWaypointRunnerProfileRequest) validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

type CreateWaypointRunnerProfileResponse struct{}

func (a *Activities) CreateWaypointRunnerProfile(ctx context.Context, req CreateWaypointRunnerProfileRequest) (CreateWaypointRunnerProfileResponse, error) {
	var resp CreateWaypointRunnerProfileResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("failed to validate request: %w", err)
	}

	client, err := a.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	if err := a.createWaypointRunnerProfile(ctx, client, req); err != nil {
		return resp, fmt.Errorf("failed to create runner profile: %w", err)
	}

	return resp, nil
}

type waypointRunnerProfileCreator interface {
	createWaypointRunnerProfile(context.Context, waypointClientODRConfigUpserter, CreateWaypointRunnerProfileRequest) error
}

type waypointClientODRConfigUpserter interface {
	UpsertOnDemandRunnerConfig(context.Context, *gen.UpsertOnDemandRunnerConfigRequest, ...grpc.CallOption) (*gen.UpsertOnDemandRunnerConfigResponse, error)
}

var _ waypointRunnerProfileCreator = (*wpRunnerProfileCreator)(nil)

type wpRunnerProfileCreator struct{}

func (w *wpRunnerProfileCreator) createWaypointRunnerProfile(ctx context.Context, client waypointClientODRConfigUpserter, req CreateWaypointRunnerProfileRequest) error {
	odrServiceAccount := runnerOdrServiceAccountName(req.InstallID)

	_, err := client.UpsertOnDemandRunnerConfig(ctx, &gen.UpsertOnDemandRunnerConfigRequest{
		Config: &gen.OnDemandRunnerConfig{
			Name:   req.InstallID,
			OciUrl: defaultODRImageURL,
			TargetRunner: &gen.Ref_Runner{
				Target: &gen.Ref_Runner_Id{
					Id: &gen.Ref_RunnerId{
						Id: req.InstallID,
					},
				},
			},
			PluginType: "kubernetes",
			PluginConfig: []byte(fmt.Sprintf(`{
"service_account": "%s",
"image_pull_policy": "%s"
}`, odrServiceAccount, defaultODRImagePullPolicy)),
			ConfigFormat:         gen.Hcl_JSON,
			Default:              false,
			EnvironmentVariables: map[string]string{},
		},
	})

	return err
}
