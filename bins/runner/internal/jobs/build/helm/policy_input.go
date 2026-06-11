package helm

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
	"sigs.k8s.io/yaml"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/helm"
)

var nuonTemplatePattern = regexp.MustCompile(`{{[^}]*\.nuon[^}]*}}`)
var nuonTemplatePlaceholder = "__NUON_VALUE__"

func (h *handler) buildPolicyInput(ctx context.Context, l *zap.Logger) ([]AdmissionReviewInput, error) {
	if h.state == nil || h.state.cfg == nil {
		l.Debug("policy input skipped: missing state or config")
		return nil, nil
	}

	chartPath := h.state.chartPath
	if chartPath == "" {
		l.Debug("policy input skipped: chart path missing")
		return nil, nil
	}
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load chart")
	}

	values, err := helm.ChartValues(h.state.cfg.ValuesFiles, h.state.cfg.Values, "")
	if err != nil {
		return nil, fmt.Errorf("unable to load helm values: %w", err)
	}
	l.Debug("policy input helm values loaded", zap.Int("values_files_count", len(h.state.cfg.ValuesFiles)), zap.Int("values_count", len(values)))
	values = sanitizePolicyValues(values)

	policyInputs, err := toPolicyAdmissionInputs(chart, values)
	if err != nil {
		return nil, err
	}

	if len(policyInputs) == 0 {
		l.Warn("no helm policy inputs generated")
		return nil, nil
	}

	h.state.policyInput = policyInputs
	return policyInputs, nil
}

func toPolicyAdmissionInputs(chart *chart.Chart, values map[string]interface{}) ([]AdmissionReviewInput, error) {
	if chart == nil {
		return nil, nil
	}

	values = sanitizePolicyValues(values)

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

func sanitizePolicyValues(values map[string]interface{}) map[string]interface{} {
	if len(values) == 0 {
		return values
	}

	sanitizePolicyValue(values)
	return values
}

func sanitizePolicyValue(value interface{}) interface{} {
	switch typedValue := value.(type) {
	case map[string]interface{}:
		for key, item := range typedValue {
			typedValue[key] = sanitizePolicyValue(item)
		}
		return typedValue
	case map[interface{}]interface{}:
		for key, item := range typedValue {
			typedValue[key] = sanitizePolicyValue(item)
		}
		return typedValue
	case []interface{}:
		for idx, item := range typedValue {
			typedValue[idx] = sanitizePolicyValue(item)
		}
		return typedValue
	case string:
		if strings.Contains(typedValue, "{{") && strings.Contains(typedValue, ".nuon") {
			return nuonTemplatePattern.ReplaceAllString(typedValue, nuonTemplatePlaceholder)
		}
		return typedValue
	default:
		return value
	}
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
	copy := make(map[string]interface{}, len(obj))
	for key, value := range obj {
		copy[key] = value
	}

	return copy
}
