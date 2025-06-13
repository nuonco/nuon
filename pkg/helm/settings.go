package helm

import "helm.sh/helm/v4/pkg/cli"

func LoadEnvSettings() (*cli.EnvSettings, error) {
	return cli.New(), nil
}
