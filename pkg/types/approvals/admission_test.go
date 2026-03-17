package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMultiDocYAMLToAdmissionReviews(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []AdmissionReviewInput
		expectError bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []AdmissionReviewInput{},
		},
		{
			name:     "whitespace only",
			input:    "   \n\t  ",
			expected: []AdmissionReviewInput{},
		},
		{
			name: "single document - Deployment",
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: default
spec:
  replicas: 3`,
			expected: []AdmissionReviewInput{
				{
					Review: AdmissionReviewRequest{
						Kind: AdmissionReviewKind{
							Kind:    "Deployment",
							Group:   "apps",
							Version: "v1",
						},
						Object: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"name":      "nginx",
								"namespace": "default",
							},
							"spec": map[string]interface{}{
								"replicas": float64(3),
							},
						},
					},
				},
			},
		},
		{
			name: "single document - core API (no group)",
			input: `apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: LoadBalancer`,
			expected: []AdmissionReviewInput{
				{
					Review: AdmissionReviewRequest{
						Kind: AdmissionReviewKind{
							Kind:    "Service",
							Group:   "",
							Version: "v1",
						},
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"name": "my-service",
							},
							"spec": map[string]interface{}{
								"type": "LoadBalancer",
							},
						},
					},
				},
			},
		},
		{
			name: "multi-document YAML",
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  key: value
---
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: ClusterIP`,
			expected: []AdmissionReviewInput{
				{
					Review: AdmissionReviewRequest{
						Kind: AdmissionReviewKind{
							Kind:    "ConfigMap",
							Group:   "",
							Version: "v1",
						},
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name": "config",
							},
							"data": map[string]interface{}{
								"key": "value",
							},
						},
					},
				},
				{
					Review: AdmissionReviewRequest{
						Kind: AdmissionReviewKind{
							Kind:    "Service",
							Group:   "",
							Version: "v1",
						},
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"name": "my-service",
							},
							"spec": map[string]interface{}{
								"type": "ClusterIP",
							},
						},
					},
				},
			},
		},
		{
			name: "multi-document with empty documents",
			input: `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
---
---
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
---`,
			expected: []AdmissionReviewInput{
				{
					Review: AdmissionReviewRequest{
						Kind: AdmissionReviewKind{
							Kind:    "ConfigMap",
							Group:   "",
							Version: "v1",
						},
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name": "config",
							},
						},
					},
				},
				{
					Review: AdmissionReviewRequest{
						Kind: AdmissionReviewKind{
							Kind:    "Secret",
							Group:   "",
							Version: "v1",
						},
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"name": "my-secret",
							},
						},
					},
				},
			},
		},
		{
			name: "missing kind and apiVersion",
			input: `metadata:
  name: unknown
data:
  key: value`,
			expected: []AdmissionReviewInput{
				{
					Review: AdmissionReviewRequest{
						Kind: AdmissionReviewKind{
							Kind:    "",
							Group:   "",
							Version: "",
						},
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"name": "unknown",
							},
							"data": map[string]interface{}{
								"key": "value",
							},
						},
					},
				},
			},
		},
		{
			name:        "invalid YAML",
			input:       "invalid: yaml: content: [",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMultiDocYAMLToAdmissionReviews(tt.input)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result), "number of results should match")

			for i := range tt.expected {
				assert.Equal(t, tt.expected[i].Review.Kind, result[i].Review.Kind, "kind should match for document %d", i)
				assert.Equal(t, tt.expected[i].Review.Object, result[i].Review.Object, "object should match for document %d", i)
			}
		})
	}
}

func TestExtractKindInfo(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected AdmissionReviewKind
	}{
		{
			name: "apps/v1 Deployment",
			obj: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
			},
			expected: AdmissionReviewKind{
				Kind:    "Deployment",
				Group:   "apps",
				Version: "v1",
			},
		},
		{
			name: "core v1 Service",
			obj: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
			},
			expected: AdmissionReviewKind{
				Kind:    "Service",
				Group:   "",
				Version: "v1",
			},
		},
		{
			name: "networking.k8s.io/v1 Ingress",
			obj: map[string]interface{}{
				"apiVersion": "networking.k8s.io/v1",
				"kind":       "Ingress",
			},
			expected: AdmissionReviewKind{
				Kind:    "Ingress",
				Group:   "networking.k8s.io",
				Version: "v1",
			},
		},
		{
			name:     "empty object",
			obj:      map[string]interface{}{},
			expected: AdmissionReviewKind{},
		},
		{
			name: "only kind",
			obj: map[string]interface{}{
				"kind": "Pod",
			},
			expected: AdmissionReviewKind{
				Kind: "Pod",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractKindInfo(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}
