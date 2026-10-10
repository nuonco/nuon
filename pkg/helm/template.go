package helm

import (
	"fmt"

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

// RenderManifest renders a chart on the client, using the release name and
// namespace Helm will install into. It does not contact the cluster.
func RenderManifest(ch *chart.Chart, values map[string]interface{}, releaseName, namespace string) (string, error) {
	if ch == nil {
		return "", nil
	}
	if releaseName == "" {
		return "", fmt.Errorf("release name is required to render the chart")
	}
	if namespace == "" {
		namespace = "default"
	}
	if values == nil {
		values = map[string]interface{}{}
	}

	client := action.NewInstall(&action.Configuration{})
	client.DryRun = true
	client.ClientOnly = true
	client.IncludeCRDs = true
	client.Namespace = namespace
	client.ReleaseName = releaseName

	rel, err := client.Run(ch, values)
	if err != nil {
		return "", errors.Wrap(err, "unable to render helm templates")
	}
	return rel.Manifest, nil
}
