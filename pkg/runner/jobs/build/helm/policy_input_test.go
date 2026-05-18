package helm

import (
	"testing"

	"github.com/stretchr/testify/require"
	"helm.sh/helm/v4/pkg/chart/common"
	chart "helm.sh/helm/v4/pkg/chart/v2"
)

func TestToPolicyAdmissionInputsNilChart(t *testing.T) {
	inputs, err := toPolicyAdmissionInputs(nil, nil)
	require.NoError(t, err)
	require.Nil(t, inputs)
}

func TestToPolicyAdmissionInputsEmptyManifest(t *testing.T) {
	ch := testChart(`{{- if .Values.enabled }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
{{- end }}`)

	inputs, err := toPolicyAdmissionInputs(ch, map[string]interface{}{"enabled": false})
	require.NoError(t, err)
	require.Nil(t, inputs)
}

func TestToPolicyAdmissionInputsInvalidYaml(t *testing.T) {
	ch := testChart("kind: [invalid")

	inputs, err := toPolicyAdmissionInputs(ch, nil)
	require.Error(t, err)
	require.Nil(t, inputs)
}

func TestToPolicyAdmissionInputsParsesManifests(t *testing.T) {
	ch := testChart(`apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Values.configName }}
  namespace: default
  labels:
    app: demo
  annotations:
    foo: bar
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
  extra: ignored
spec: {}`)

	inputs, err := toPolicyAdmissionInputs(ch, map[string]interface{}{"configName": "config"})
	require.NoError(t, err)
	require.Len(t, inputs, 2)

	first := inputs[0].Review
	require.Equal(t, AdmissionReviewKind{Kind: "ConfigMap", Version: "v1"}, first.Kind)
	require.Equal(t, map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":        "config",
			"namespace":   "default",
			"labels":      map[string]interface{}{"app": "demo"},
			"annotations": map[string]interface{}{"foo": "bar"},
		},
	}, first.Object)

	second := inputs[1].Review
	require.Equal(t, AdmissionReviewKind{Kind: "Deployment", Group: "apps", Version: "v1"}, second.Kind)
	require.Equal(t, map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"extra": "ignored",
			"name":  "deploy",
		},
		"spec": map[string]interface{}{},
	}, second.Object)
}

func TestToPolicyAdmissionInputsSanitizesNuonValues(t *testing.T) {
	ch := testChart(`apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Values.name }}
data:
  message: {{ tpl .Values.message . }}
`)

	values := map[string]interface{}{
		"name":    "config",
		"message": "{{ .nuon.inputs.foo }}",
	}

	inputs, err := toPolicyAdmissionInputs(ch, values)
	require.NoError(t, err)
	require.Len(t, inputs, 1)
}

func TestExtractKindInfo(t *testing.T) {
	kind := extractKindInfo(map[string]interface{}{
		"kind":       "Service",
		"apiVersion": "v1",
	})
	require.Equal(t, AdmissionReviewKind{Kind: "Service", Version: "v1"}, kind)

	kind = extractKindInfo(map[string]interface{}{
		"kind":       "Deployment",
		"apiVersion": "apps/v1",
	})
	require.Equal(t, AdmissionReviewKind{Kind: "Deployment", Group: "apps", Version: "v1"}, kind)
}

func TestExtractMetadataObject(t *testing.T) {
	obj := extractMetadataObject(map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":        "test",
			"namespace":   "default",
			"labels":      map[string]interface{}{"app": "demo"},
			"annotations": map[string]interface{}{"foo": "bar"},
			"ignored":     "value",
		},
		"spec": map[string]interface{}{"replicas": 1},
	})
	require.Equal(t, map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":        "test",
			"namespace":   "default",
			"labels":      map[string]interface{}{"app": "demo"},
			"annotations": map[string]interface{}{"foo": "bar"},
			"ignored":     "value",
		},
		"spec": map[string]interface{}{"replicas": 1},
	}, obj)
}

func testChart(template string) *chart.Chart {
	return &chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion: chart.APIVersionV2,
			Name:       "test",
			Version:    "0.1.0",
		},
		Templates: []*common.File{
			{
				Name: "templates/test.yaml",
				Data: []byte(template),
			},
		},
	}
}
