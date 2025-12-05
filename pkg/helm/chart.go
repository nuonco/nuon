package helm

import (
	"fmt"
	"net/url"
	"strings"

	"helm.sh/helm/v4/pkg/action"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
	"helm.sh/helm/v4/pkg/cli"
)

// resolveChartName returns the proper repository and name values that
// the ChartPathOptions need. This is copied from Terraform.
func ResolveChartName(repository, name string) (string, string, error) {
	_, err := url.ParseRequestURI(repository)
	if err == nil {
		return repository, name, nil
	}

	if strings.Index(name, "/") == -1 && repository != "" {
		name = fmt.Sprintf("%s/%s", repository, name)
	}

	return "", name, nil
}

func GetChartByPath(urlOrPath string) (*chart.Chart, error) {
	c, err := loader.Load(urlOrPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get chart: %w", err)
	}

	return c, nil
}

func GetChart(name string, cpo *action.ChartPathOptions, settings *cli.EnvSettings) (*chart.Chart, string, error) {
	path, err := cpo.LocateChart(name, settings)
	if err != nil {
		return nil, "", err
	}

	c, err := loader.Load(path)
	if err != nil {
		return nil, "", err
	}

	return c, path, nil
}

func ChartPathOptions(repository, chart, version string) (*action.ChartPathOptions, string, error) {
	repositoryURL, chartName, err := ResolveChartName(
		repository, strings.TrimSpace(chart))
	if err != nil {
		return nil, "", err
	}

	// Determine our version string
	if version == "" {
		version = ">0.0.0-0"
	}
	version = strings.TrimSpace(version)

	// Initialize our chart options
	return &action.ChartPathOptions{
		RepoURL: repositoryURL,
		Version: version,
	}, chartName, nil
}
