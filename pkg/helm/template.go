package helm

import (
	"github.com/pkg/errors"
	"helm.sh/helm/v4/pkg/action"
	chart "helm.sh/helm/v4/pkg/chart/v2"
)

func TemplateChart(chart *chart.Chart, values map[string]interface{}) (string, error) {
	if chart == nil {
		return "", nil
	}

	client := action.NewInstall(&action.Configuration{})
	client.DryRun = true
	client.ClientOnly = true
	client.IncludeCRDs = true
	client.Namespace = "default"
	client.ReleaseName = "policy-input"

	rel, err := client.Run(chart, values)
	if err != nil {
		return "", errors.Wrap(err, "unable to render helm templates")
	}

	return rel.Manifest, nil
}
