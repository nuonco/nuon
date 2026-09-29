package kubernetes_manifest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakedynamic "k8s.io/client-go/dynamic/fake"
)

func Test_handler_execDelete(t *testing.T) {
	tests := map[string]struct {
		resources       []*kubernetesResource
		clientSetupFunc func() *fakedynamic.FakeDynamicClient
		expectedError   string
	}{
		"successful delete": {
			resources: []*kubernetesResource{
				{
					groupVersionResource: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
					namespace:            "test-namespace",
					name:                 "test-deployment",
					namespaced:           true,
				},
			},
			clientSetupFunc: func() *fakedynamic.FakeDynamicClient {
				client := fakedynamic.NewSimpleDynamicClient(runtime.NewScheme())

				deployment := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "apps/v1",
						"kind":       "Deployment",
						"metadata": map[string]interface{}{
							"name":      "test-deployment",
							"namespace": "test-namespace",
						},
					},
				}
				gvr := schema.GroupVersionResource{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				}
				_, _ = client.Resource(gvr).Namespace("test-namespace").Create(
					context.TODO(),
					deployment,
					metav1.CreateOptions{},
				)

				return client
			},
			expectedError: "",
		},
		"delete error resource not found": {
			resources: []*kubernetesResource{
				{
					groupVersionResource: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
					namespace:            "default",
					name:                 "test-deployment",
				},
			},
			clientSetupFunc: func() *fakedynamic.FakeDynamicClient {
				client := fakedynamic.NewSimpleDynamicClient(runtime.NewScheme())
				return client
			},
			expectedError: "delete error for resource [ default/test-deployment]: deployments.apps \"test-deployment\" not found",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.clientSetupFunc()
			h := &handler{}
			out, err := h.execDelete(context.Background(), client, test.resources, false)
			if test.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, len(test.resources), len(*out), "expected number of resources to be deleted should match the input resources")
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}
