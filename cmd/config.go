package cmd

import (
	"github.com/powertoolsdev/go-common/config"
	instance "github.com/powertoolsdev/workers-instances/internal"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "default")

	// instance defaults
	config.RegisterDefault("instance.bucket", "nuon-installations-stage")
	config.RegisterDefault("instance.bucket_region", "us-west-2")
	config.RegisterDefault("instance.role_arn", "arn:aws:iam::618886478608:role/install-k8s-admin-stage")
	config.RegisterDefault("instance.waypoint_token_secret_namespace", "default")
	config.RegisterDefault("instance.waypoint_server_root_domain", "orgs-stage.nuon.co")
}

type Config struct {
	config.Base       `config:",squash"`
	TemporalHost      string `config:"temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`

	// Domain specific configs
	Cfg instance.Config `config:"instance"`
}
