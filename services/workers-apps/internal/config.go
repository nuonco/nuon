package internal

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/services/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_namespace", "apps")

	//org bucket default
	config.RegisterDefault("org_orgs_bucket_name", "nuon-orgs-stage")
}

type Config struct {
	worker.Config `config:",squash"`

	OrgsEcrAccessRoleArn string `config:"orgs_ecr_access_role_arn" validate:"required" json:"orgs_ecr_access_iam_role_arn"`
	OrgsRoleTemplate     string `config:"orgs_role_template" validate:"required"`
	OrgsBucketName       string `config:"orgs_bucket_name" json:"org_orgs_bucket_name" validate:"required"`

	WaypointTokenNamespace   string `config:"waypoint_token_namespace" json:"waypoint_token_namespace" validate:"required"`
	WaypointServerRootDomain string `config:"waypoint_server_root_domain" json:"waypoint_server_root_domain" validate:"required"`

	// authenticate with orgs cluster
	OrgsK8sCAData         string `config:"orgs_k8s_ca_data"`
	OrgsK8sPublicEndpoint string `config:"orgs_k8s_public_endpoint"`
	OrgsK8sClusterID      string `config:"orgs_k8s_cluster_id"`
	OrgsK8sRoleArn        string `config:"orgs_k8s_role_arn"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
