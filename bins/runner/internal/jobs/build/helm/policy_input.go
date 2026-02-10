package helm

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
	"sigs.k8s.io/yaml"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/helm"
)

func (h *handler) buildPolicyInput(ctx context.Context, l *zap.Logger) ([]AdmissionReviewInput, error) {
	if h.state == nil || h.state.cfg == nil {
		return nil, nil
	}

	chartPath := h.state.chartPath
	if chartPath == "" {
		return nil, nil
	}
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load chart")
	}

	values, err := helm.ChartValues(h.state.cfg.ValuesFiles, h.state.cfg.Values)
	if err != nil {
		return nil, fmt.Errorf("unable to load helm values: %w", err)
	}

	policyInputs, err := toPolicyAdmissionInputs(chart, values)
	if err != nil {
		return nil, err
	}

	if len(policyInputs) == 0 {
		l.Debug("no helm policy inputs generated")
		return nil, nil
	}

	h.state.policyInput = policyInputs
	return policyInputs, nil
}

func toPolicyAdmissionInputs(chart *chart.Chart, values map[string]interface{}) ([]AdmissionReviewInput, error) {
	if chart == nil {
		return nil, nil
	}

	manifests, err := helm.TemplateChart(chart, values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render helm templates")
	}

	if strings.TrimSpace(manifests) == "" {
		return nil, nil
	}

	docs := strings.Split(manifests, "---")
	inputs := make([]AdmissionReviewInput, 0, len(docs))
	for _, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var obj map[string]interface{}
		if err := yaml.Unmarshal([]byte(doc), &obj); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal helm manifest")
		}

		if len(obj) == 0 {
			continue
		}

		inputs = append(inputs, AdmissionReviewInput{
			Review: AdmissionReviewRequest{
				Kind:   extractKindInfo(obj),
				Object: extractMetadataObject(obj),
			},
		})
	}

	return inputs, nil
}

func extractKindInfo(obj map[string]interface{}) AdmissionReviewKind {
	kind := AdmissionReviewKind{}

	if k, ok := obj["kind"].(string); ok {
		kind.Kind = k
	}

	if apiVersion, ok := obj["apiVersion"].(string); ok {
		parts := strings.Split(apiVersion, "/")
		if len(parts) == 2 {
			kind.Group = parts[0]
			kind.Version = parts[1]
		} else if len(parts) == 1 {
			kind.Version = parts[0]
		}
	}

	return kind
}

func extractMetadataObject(obj map[string]interface{}) map[string]interface{} {
	metadata := map[string]interface{}{}
	if raw, ok := obj["metadata"].(map[string]interface{}); ok {
		for _, key := range []string{"name", "namespace", "labels", "annotations"} {
			if value, ok := raw[key]; ok {
				metadata[key] = value
			}
		}
	}

	return map[string]interface{}{
		"metadata": metadata,
	}
}
