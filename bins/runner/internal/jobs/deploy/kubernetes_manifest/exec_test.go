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

// func Test_handler_execApply(t *testing.T) {
// 	tests := map[string]struct {
// 		resources       []*kubernetesResource
// 		clientSetupFunc func() *fakedynamic.FakeDynamicClient
// 		expectedError   string
// 	}{
// 		"successful apply, existing and new resource": {
// 			resources: []*kubernetesResource{
// 				{
// 					groupVersionResource: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
// 					namespace:            "default",
// 					name:                 "test-deployment",
// 					obj: &unstructured.Unstructured{
// 						Object: map[string]interface{}{
// 							"apiVersion": "apps/v1",
// 							"kind":       "Deployment",
// 							"metadata": map[string]interface{}{
// 								"name":      "test-deployment",
// 								"namespace": "default",
// 								"labels": map[string]interface{}{
// 									"app": "test-app",
// 								},
// 							},
// 						},
// 					},
// 				},
// 				{
// 					groupVersionResource: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
// 					namespace:            "default",
// 					name:                 "test-deployment-2",
// 					obj: &unstructured.Unstructured{
// 						Object: map[string]interface{}{
// 							"apiVersion": "apps/v1",
// 							"kind":       "Deployment",
// 							"metadata": map[string]interface{}{
// 								"name":      "test-deployment-2",
// 								"namespace": "default",
// 							},
// 						},
// 					},
// 				},
// 			},
// 			clientSetupFunc: func() *fakedynamic.FakeDynamicClient {

// 				client := fakedynamic.NewSimpleDynamicClient(runtime.NewScheme())

// 				// Preload the resource into the fake client
// 				deployment := &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "apps/v1",
// 						"kind":       "Deployment",
// 						"metadata": map[string]interface{}{
// 							"name":      "test-deployment",
// 							"namespace": "default",
// 						},
// 					},
// 				}
// 				gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
// 				_, _ = client.Resource(gvr).Namespace("default").Create(context.TODO(), deployment, metav1.CreateOptions{})

// 				client.Fake.PrependReactor("patch", "deployments", func(action kubetesting.Action) (bool, runtime.Object, error) {
// 					patch := action.(kubetesting.PatchAction)
// 					u := &unstructured.Unstructured{}
// 					if err := json.Unmarshal(patch.GetPatch(), &u.Object); err != nil {
// 						return true, nil, err
// 					}
// 					return true, u, nil
// 				})

// 				return client

// 				// Simulate the Apply operation by creating the resource if it does not exist
// 				// client.PrependReactor("apply", "deployments", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
// 				// 	patchAction := action.(kubetesting.UpdateAction)
// 				// 	// Extract the object from the action
// 				// 	obj := patchAction.GetObject().(*unstructured.Unstructured)
// 				// 	gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

// 				// 	// Check if the resource exists
// 				// 	existing, err := client.Resource(gvr).Namespace(obj.GetNamespace()).Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
// 				// 	if err != nil {
// 				// 		// If the resource does not exist, create it
// 				// 		return true, obj, nil
// 				// 	}
// 				// 	return true, existing, nil
// 				// })

// 			},
// 			expectedError: "",
// 		},
// 		// "apply error": {
// 		// 	resources: []*kubernetesResource{
// 		// 		{
// 		// 			groupVersionResource: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
// 		// 			namespace:            "default",
// 		// 			name:                 "test-deployment",
// 		// 			obj:                  &unstructured.Unstructured{},
// 		// 		},
// 		// 	},
// 		// 	clientSetupFunc: func() *fakedynamic.FakeDynamicClient {
// 		// 		client := fakedynamic.NewSimpleDynamicClient(runtime.NewScheme())
// 		// 		client.PrependReactor("apply", "deployments", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
// 		// 			return true, nil, fmt.Errorf("mock apply error")
// 		// 		})
// 		// 		return client
// 		// 	},
// 		// 	expectedError: "apply error for resource [Group: apps, Version: v1, Kind: , Namespace: default, Name: test-deployment]: mock apply error",
// 		// },
// 	}

// 	for name, test := range tests {
// 		t.Run(name, func(t *testing.T) {
// 			client := test.clientSetupFunc()
// 			h := &handler{}
// 			err := h.execApply(context.Background(), client, test.resources)
// 			if test.expectedError == "" {
// 				assert.NoError(t, err)
// 			} else {
// 				assert.EqualError(t, err, test.expectedError)
// 			}
// 		})
// 	}
// }

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

				// Preload the resource into the fake client
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
