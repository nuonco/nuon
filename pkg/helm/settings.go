package helm

import "helm.sh/helm/v3/pkg/cli"

func LoadEnvSettings() (*cli.EnvSettings, error) {
	return cli.New(), nil
}
