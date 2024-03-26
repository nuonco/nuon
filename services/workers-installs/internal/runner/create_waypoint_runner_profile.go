package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/kube"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	waypoint "github.com/powertoolsdev/mono/pkg/waypoint/client"
	"github.com/powertoolsdev/mono/pkg/waypoint/client/k8s"
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
	ClusterInfo          kube.ClusterInfo `json:"cluster_info" validate:"required"`
	RunnerType           installsv1.RunnerType

	// additional information for ecs
	LogGroupName   string
	EcsClusterInfo *runnerv1.ECSClusterInfo

	// additional information for aws eks
	AwsRegion string `json:"aws_region"`
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
	if err != nil {
		return resp, fmt.Errorf("unable to get org provider: %w", err)
	}

	client, err := provider.Fetch(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to get client: %w", err)
	}

	if err := a.createWaypointRunnerProfile(ctx, client, req); err != nil {
		return resp, fmt.Errorf("failed to create runner profile: %w", err)
	}

	return resp, nil
}

func (a *Activities) roleNameFromARN(roleARN string) (string, error) {
	pieces := strings.SplitN(roleARN, "role/", 2)
	if len(pieces) != 2 {
		return "", fmt.Errorf("unable to parse name from role arn")
	}

	return pieces[1], nil
}

func (a *Activities) createWaypointRunnerProfile(ctx context.Context, client gen.WaypointClient, req CreateWaypointRunnerProfileRequest) error {
	odrServiceAccount := runnerOdrServiceAccountName(req.InstallID)

	pluginCfg := map[string]interface{}{}
	cfgReq := &gen.UpsertOnDemandRunnerConfigRequest{
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
			ConfigFormat:         gen.Hcl_JSON,
			Default:              false,
			EnvironmentVariables: map[string]string{},
		},
	}

	switch req.RunnerType {
	case installsv1.RunnerType_RUNNER_TYPE_AWS_ECS:
		// NOTE(jm): since waypoint requires role-names, instead of ARNs, but we require ARNs in other places,
		// we just parse out the name here.
		runnerRoleName, err := a.roleNameFromARN(req.EcsClusterInfo.RunnerIamRoleArn)
		if err != nil {
			return fmt.Errorf("invalid runner arn: %w", err)
		}
		odrRoleName, err := a.roleNameFromARN(req.EcsClusterInfo.OdrIamRoleArn)
		if err != nil {
			return fmt.Errorf("invalid runner arn: %w", err)
		}

		cfgReq.Config.PluginType = "aws-ecs"
		pluginCfg["log_group"] = req.LogGroupName
		pluginCfg["execution_role_name"] = runnerRoleName
		pluginCfg["task_role_name"] = odrRoleName
		pluginCfg["cluster"] = req.EcsClusterInfo.ClusterName
		pluginCfg["region"] = req.AwsRegion
		pluginCfg["odr_cpu"] = "512"
		pluginCfg["odr_memory"] = "2048"
		pluginCfg["security_group_id"] = req.EcsClusterInfo.SecurityGroupId
		pluginCfg["subnets"] = strings.Join(req.EcsClusterInfo.SubnetIds, ",")
	case installsv1.RunnerType_RUNNER_TYPE_AWS_EKS:
		cfgReq.Config.PluginType = "kubernetes"
		pluginCfg["service_account"] = odrServiceAccount
		pluginCfg["image_pull_policy"] = defaultODRImagePullPolicy
		cfgReq.Config.EnvironmentVariables["AWS_REGION_DEFAULT"] = req.AwsRegion
	case installsv1.RunnerType_RUNNER_TYPE_AZURE_AKS:
		cfgReq.Config.PluginType = "kubernetes"
		pluginCfg["service_account"] = odrServiceAccount
		pluginCfg["image_pull_policy"] = defaultODRImagePullPolicy
	default:
		return fmt.Errorf("unsupported runner type")
	}

	cfgJson, err := json.MarshalIndent(pluginCfg, "", "\t")
	if err != nil {
		return fmt.Errorf("unable to marshal runner plugin config: %w", err)
	}
	cfgReq.Config.PluginConfig = cfgJson

	_, err = client.UpsertOnDemandRunnerConfig(ctx, cfgReq)
	return err
}
