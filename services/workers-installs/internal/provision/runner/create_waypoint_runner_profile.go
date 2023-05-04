package runner

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/kube"
	waypoint "github.com/powertoolsdev/mono/pkg/waypoint/client"
	"google.golang.org/grpc"
)

const (
	defaultODRImagePullPolicy string = "Always"
	defaultODRImageURL        string = "public.ecr.aws/p7e3r5y0/waypoint-odr:v0.0.5"
)

type CreateWaypointRunnerProfileRequest struct {
	TokenSecretNamespace string           `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string           `json:"org_server_address" validate:"required"`
	OrgID                string           `json:"org_id" validate:"required"`
	InstallID            string           `json:"install_id" validate:"required"`
	AwsRegion            string           `json:"aws_region" validate:"required"`
	ClusterInfo          kube.ClusterInfo `json:"cluster_info" validate:"required"`
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

	provider, err := waypoint.NewK8sProvider(a.v, waypoint.WithConfig(waypoint.Config{
		Address: req.OrgServerAddr,
		Token: waypoint.Token{
			Namespace: req.TokenSecretNamespace,
			Name:      waypoint.DefaultTokenSecretName(req.OrgID),
			Key:       waypoint.DefaultTokenSecretKey,
		},
		ClusterInfo: &req.ClusterInfo,
	}))
	if err != nil {
		return resp, fmt.Errorf("unable to get org provider: %w", err)
	}
	if err != nil {
		return resp, fmt.Errorf("unable to get org provider: %w", err)
	}

	client, err := provider.GetClient(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to get client: %w", err)
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
			ConfigFormat: gen.Hcl_JSON,
			Default:      false,
			EnvironmentVariables: map[string]string{
				"AWS_REGION_DEFAULT": req.AwsRegion,
			},
		},
	})

	return err
}
