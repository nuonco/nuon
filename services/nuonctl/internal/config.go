package internal

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/common/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("env", "local")

	// temporal defaults
	config.RegisterDefault("dev_temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "default")
	config.RegisterDefault("stage_temporal_host", "temporal-frontend.nuon.us-west-2.stage.nuon.cloud:7233")

	// github defaults
	config.RegisterDefault("github_app_id", "261597")

	// buckets
	config.RegisterDefault("installs_bucket", "nuon-org-installations-stage")
	config.RegisterDefault("deployments_bucket", "nuon-org-deployments-stage")
	config.RegisterDefault("orgs_bucket", "nuon-orgs-stage")

	config.RegisterDefault("support_iam_role_arn", "arn:aws:iam::766121324316:role/nuon-internal-support-stage")
}

type Config struct {
	config.Base `config:",squash"`

	// temporal configurations
	DevTemporalHost   string `config:"dev_temporal_host"`
	StageTemporalHost string `config:"stage_temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`

	// github configurations
	GithubAppID  string `config:"github_app_id"`
	GithubAppKey string `config:"github_app_key"`

	// buckets
	InstallsBucket    string `config:"installs_bucket"`
	OrgsBucket        string `config:"orgs_bucket"`
	DeploymentsBucket string `config:"deployments_bucket"`

	// aws config
	SupportIAMRoleArn string `config:"support_iam_role_arn"`
}

func (c Config) Validate(v *validator.Validate) error {
	return v.Struct(c)
}
