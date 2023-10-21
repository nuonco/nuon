package terraformcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type OrgsK8sOutputs struct {
	AccessRoleARNs  map[string]string `mapstructure:"access_role_arns"`
	ClusterID       string            `mapstructure:"cluster_id"`
	CAData          string            `mapstructure:"ca_data"`
	PublicEndpoint  string            `mapstructure:"public_endpoint"`
	OIDCProviderURL string            `mapstructure:"oidc_provider_url"`
	OIDCProviderARN string            `mapstructure:"oidc_provider_arn"`
}

type OrgsIAMRoleNameTemplateOutputs struct {
	DeploymentsAccess   string `mapstructure:"deployments_access"`
	SecretsAccess       string `mapstructure:"secrets_access"`
	InstallationsAccess string `mapstructure:"installations_access"`
	InstancesAccess     string `mapstructure:"instances_access"`
	OrgsAccess          string `mapstructure:"orgs_access"`
	ODRAccess           string `mapstructure:"odr_access"`
}

type OrgsECROutputs struct {
	RegistryARN string `mpastructure:"registry_arn"`
	Region      string `mpastructure:"region"`
	RegistryID  string `mpastructure:"registry_id"`
}

type OrgsWaypointOutputs struct {
	RootDomain           string `mapstructure:"root_domain"`
	TokenSecretNamespace string `mapstructure:"token_secret_namespace"`
	TokenSecretTemplate  string `mapstructure:"token_secret_template"`

	ClusterID       string `mapstructure:"cluster_id"`
	CAData          string `mapstructure:"ca_data"`
	PublicEndpoint  string `mapstructure:"public_endpoint"`
	OIDCProviderURL string `mapstructure:"oidc_provider_url"`
	OIDCProviderARN string `mapstructure:"oidc_provider_arn"`
}

type OrgsBootstrapWaypointOutputs struct {
	Domain               string `mapstructure:"domain"`
	TokenSecretNamespace string `mapstructure:"token_secret_namespace"`
	TokenSecretTemplate  string `mapstructure:"token_secret_template"`

	ClusterID       string `mapstructure:"cluster_id"`
	CAData          string `mapstructure:"ca_data"`
	PublicEndpoint  string `mapstructure:"public_endpoint"`
	OIDCProviderURL string `mapstructure:"oidc_provider_url"`
	OIDCProviderARN string `mapstructure:"oidc_provider_arn"`
}

type BucketOutputs struct {
	Name   string `mapstructure:"name"`
	Region string `mapstructure:"region"`
}

type OrgsBucketsOutputs struct {
	Deployments   BucketOutputs `mapstructure:"deployments"`
	Secrets       BucketOutputs `mapstructure:"secrets"`
	Installations BucketOutputs `mapstructure:"installations"`
	Orgs          BucketOutputs `mapstructure:"orgs"`
}

type IAMRoleOutputs struct {
	Description string `mapstructure:"description"`
	ARN         string `mapstructure:"arn"`
}

type OrgsIAMRolesOutputs struct {
	InstallK8sAccess IAMRoleOutputs `mapstructure:"install_k8s_access"`
	Support          IAMRoleOutputs `mapstructure:"support"`
}

type OrgsAccountOutputs struct {
	ID string `mapstructure:"id"`
}

type OrgsPublicDomainOutputs struct {
	Nameservers []string `mapstructure:"nameservers"`
	Domain      string   `mapstructure:"domain"`
	ZoneID      string   `mapstructure:"zone_id"`
}

type OrgsOutputs struct {
	K8s                            OrgsK8sOutputs                 `mapstructure:"k8s"`
	OrgsIAMRoleNameTemplateOutputs OrgsIAMRoleNameTemplateOutputs `mapstructure:"org_iam_role_name_templates"`
	ECR                            OrgsECROutputs                 `mapstructure:"ecr"`
	Waypoint                       OrgsWaypointOutputs            `mapstructure:"waypoint"`
	BootstrapWaypoint              OrgsBootstrapWaypointOutputs   `mapstructure:"bootstrap_waypoint"`
	Buckets                        OrgsBucketsOutputs             `mapstructure:"buckets"`
	IAMRoles                       OrgsIAMRolesOutputs            `mapstructure:"iam_roles"`
	Account                        OrgsAccountOutputs             `mapstructure:"accounts"`
	PublicDomain                   OrgsPublicDomainOutputs        `mapstructure:"public_domain"`
}

func (i *OrgsOutputs) parse(outputs []*tfe.StateVersionOutput) error {
	obj := make(map[string]interface{})
	for _, out := range outputs {
		obj[out.Name] = out.Value
	}

	if err := mapstructure.Decode(obj, i); err != nil {
		return fmt.Errorf("unable to parse outputs: %w", err)
	}

	return nil
}

func NewOrgsOutputs(client *tfe.Client, lc fx.Lifecycle, cfg *internal.Config, l *zap.Logger) (*OrgsOutputs, error) {
	out := &OrgsOutputs{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			l.Info("fetching terraform cloud outputs", zap.String("workspace", cfg.TFEOrgsWorkspaceID))
			outputs, err := client.StateVersionOutputs.ReadCurrent(ctx, cfg.TFEOrgsWorkspaceID)
			if err != nil {
				return fmt.Errorf("unable to get outputs: %w", err)
			}

			if err := out.parse(outputs.Items); err != nil {
				return fmt.Errorf("unable to parse outputs: %w", err)
			}

			return nil
		},
	})
	return out, nil
}
