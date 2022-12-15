package cmd

import "github.com/powertoolsdev/go-common/config"

func init() { //nolint: gochecknoinits
	config.RegisterDefault("http_port", "8080")
	config.RegisterDefault("http_address", "0.0.0.0")
	config.RegisterDefault("db_user", "postgres")
	config.RegisterDefault("db_password", "postgres")
	config.RegisterDefault("db_port", "5432")
	config.RegisterDefault("db_name", "api")
	config.RegisterDefault("db_ssl_mode", "disable")
	config.RegisterDefault("db_host", "localhost")
	config.RegisterDefault("db_use_zap", false)
	config.RegisterDefault("db_use_iam", false)
	config.RegisterDefault("db_region", "us-west-2")
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "default")

	config.RegisterDefault("github_app_id", "261597")
}

type Config struct {
	config.Base       `config:",squash"`
	HTTPPort          string `config:"http_port"`
	HTTPAddress       string `config:"http_address"`
	DBName            string `config:"db_name"`
	DBHost            string `config:"db_host"`
	DBPort            string `config:"db_port"`
	DBSSLMode         string `config:"db_ssl_mode"`
	DBPassword        string `config:"db_password"`
	DBUser            string `config:"db_user"`
	DBZapLog          bool   `config:"db_use_zap"`
	DBUseIAM          bool   `config:"db_use_iam"`
	DBRegion          string `config:"db_region"`
	AuthIssuerURL     string `config:"auth0_issuer_url"`
	AuthAudience      string `config:"auth0_audience"`
	TemporalHost      string `config:"temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`
}
