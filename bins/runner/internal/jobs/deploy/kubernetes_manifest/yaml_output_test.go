package kubernetes_manifest

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestResourcesToMultiDocYAML(t *testing.T) {
	tests := []struct {
		name     string
		objects  []*unstructured.Unstructured
		wantDocs int
		wantErr  bool
	}{
		{
			name:     "empty slice returns empty string",
			objects:  []*unstructured.Unstructured{},
			wantDocs: 0,
			wantErr:  false,
		},
		{
			name:     "nil slice returns empty string",
			objects:  nil,
			wantDocs: 0,
			wantErr:  false,
		},
		{
			name: "single object",
			objects: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "test-cm",
							"namespace": "default",
						},
						"data": map[string]interface{}{
							"key": "value",
						},
					},
				},
			},
			wantDocs: 1,
			wantErr:  false,
		},
		{
			name: "multiple objects",
			objects: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "cm-1",
						},
					},
				},
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Secret",
						"metadata": map[string]interface{}{
							"name": "secret-1",
						},
					},
				},
				{
					Object: map[string]interface{}{
						"apiVersion": "apps/v1",
						"kind":       "Deployment",
						"metadata": map[string]interface{}{
							"name": "deploy-1",
						},
					},
				},
			},
			wantDocs: 3,
			wantErr:  false,
		},
		{
			name: "skips nil objects in slice",
			objects: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "cm-1",
						},
					},
				},
				nil,
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Secret",
						"metadata": map[string]interface{}{
							"name": "secret-1",
						},
					},
				},
			},
			wantDocs: 2,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResourcesToMultiDocYAML(tt.objects)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.wantDocs == 0 {
				assert.Empty(t, result)
				return
			}

			docs := strings.Split(result, "---\n")
			assert.Equal(t, tt.wantDocs, len(docs))

			for _, doc := range docs {
				assert.NotEmpty(t, doc)
				assert.Contains(t, doc, "apiVersion:")
				assert.Contains(t, doc, "kind:")
			}
		})
	}
}

func TestResourcesToMultiDocYAML_ContentValidation(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "my-config",
				"namespace": "my-namespace",
			},
			"data": map[string]interface{}{
				"config.yaml": "setting: true",
			},
		},
	}

	result, err := ResourcesToMultiDocYAML([]*unstructured.Unstructured{obj})
	require.NoError(t, err)

	assert.Contains(t, result, "apiVersion: v1")
	assert.Contains(t, result, "kind: ConfigMap")
	assert.Contains(t, result, "name: my-config")
	assert.Contains(t, result, "namespace: my-namespace")
	assert.Contains(t, result, "config.yaml:")
}
