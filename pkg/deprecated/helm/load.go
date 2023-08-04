package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func LoadChart(dirOrURL string) (*chart.Chart, error) {
	chrt, err := loader.Load(dirOrURL)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart %s: %w", dirOrURL, err)
	}

	return chrt, nil
}
