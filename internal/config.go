package internal

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
)

//nolint:gochecknoinits
func init() {
	// deployment defaults
	config.RegisterDefault("waypoint_token_secret_namespace", "default")
	config.RegisterDefault("waypoint_org_server_root_domain", "orgs-stage.nuon.co")

	config.RegisterDefault("http_port", "8080")
	config.RegisterDefault("http_address", "0.0.0.0")
}

type Config struct {
	config.Base `config:",squash"`

	// configs for starting and introspecting service
	GitRef      string `config:"git_ref" validate:"required"`
	HTTPPort    string `config:"http_port" validate:"required"`
	HTTPAddress string `config:"http_address" validate:"required"`

	// waypoint configuration
	WaypointTokenSecretNamespace string `config:"waypoint_token_secret_namespace" validate:"required"`
	WaypointTokenSecretTemplate  string `config:"waypoint_token_secret_template" validate:"required"`
	WaypointServerRootDomain     string `config:"waypoint_server_root_domain" validate:"required"`

	// org IAM role template names and buckets for accessing state
	DeploymentsBucket             string `config:"deployments_bucket" validate:"required"`
	OrgsDeploymentsRoleTemplate   string `config:"orgs_deployments_role_template" validate:"required"`
	OrgsInstallationsRoleTemplate string `config:"orgs_installations_role_template" validate:"required"`
	InstallationsBucket           string `config:"installations_bucket" validate:"required"`

	// configs needed to access the orgs cluster, which runs all org runner/servers
	OrgsK8sClusterID      string `config:"orgs_k8s_cluster_id" json:"orgs_k8s_cluster_id" validate:"required"`
	OrgsK8sRoleArn        string `config:"orgs_k8s_role_arn" json:"orgs_k8s_role_arn" validate:"required"`
	OrgsK8sCAData         string `config:"orgs_k8s_ca_data" json:"orgs_k8s_ca_data" validate:"required"`
	OrgsK8sPublicEndpoint string `config:"orgs_k8s_public_endpoint" json:"orgs_k8s_public_endpoint" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
