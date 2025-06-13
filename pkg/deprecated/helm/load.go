package helm

import (
	"fmt"

	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
)

func LoadChart(dirOrURL string) (*chart.Chart, error) {
	chrt, err := loader.Load(dirOrURL)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart %s: %w", dirOrURL, err)
	}

	return chrt, nil
}
