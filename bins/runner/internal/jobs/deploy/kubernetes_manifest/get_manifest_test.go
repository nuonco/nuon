package kubernetes_manifest

import (
	"testing"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/memory"
	discoveryFake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/restmapper"
	kubetesting "k8s.io/client-go/testing"
)

// func TestObjDiff(t *testing.T) {
// 	// Define test cases
// 	testCases := []struct {
// 		name         string
// 		prev         map[string]interface{}
// 		curr         map[string]interface{}
// 		expectedDiff string
// 	}{
// 		{
// 			name: "No changes",
// 			prev: map[string]interface{}{
// 				"apiVersion": "v1",
// 				"kind":       "Pod",
// 				"metadata": map[string]interface{}{
// 					"name":      "test-pod",
// 					"namespace": "default",
// 				},
// 				"spec": map[string]interface{}{
// 					"containers": []interface{}{
// 						map[string]interface{}{
// 							"name":  "nginx",
// 							"image": "nginx:1.14",
// 						},
// 					},
// 				},
// 			},
// 			curr: map[string]interface{}{
// 				"apiVersion": "v1",
// 				"kind":       "Pod",
// 				"metadata": map[string]interface{}{
// 					"name":      "test-pod",
// 					"namespace": "default",
// 				},
// 				"spec": map[string]interface{}{
// 					"containers": []interface{}{
// 						map[string]interface{}{
// 							"name":  "nginx",
// 							"image": "nginx:1.14",
// 						},
// 					},
// 				},
// 			},
// 			expectedDiff: "",
// 		},
// 		{
// 			name: "Image updated",
// 			prev: map[string]interface{}{
// 				"apiVersion": "v1",
// 				"kind":       "Pod",
// 				"metadata": map[string]interface{}{
// 					"name":      "test-pod",
// 					"namespace": "default",
// 				},
// 				"spec": map[string]interface{}{
// 					"containers": []interface{}{
// 						map[string]interface{}{
// 							"name":  "nginx",
// 							"image": "nginx:1.14",
// 						},
// 					},
// 				},
// 			},
// 			curr: map[string]interface{}{
// 				"apiVersion": "v1",
// 				"kind":       "Pod",
// 				"metadata": map[string]interface{}{
// 					"name":      "test-pod",
// 					"namespace": "default",
// 				},
// 				"spec": map[string]interface{}{
// 					"containers": []interface{}{
// 						map[string]interface{}{
// 							"name":  "nginx",
// 							"image": "nginx:1.16",
// 						},
// 					},
// 				},
// 			},
// 			expectedDiff: `
//   spec:
//     containers:
//     - image: nginx:1.14
// +   - image: nginx:1.16
// `,
// 		},
// 		{
// 			name: "Namespace added",
// 			prev: map[string]interface{}{
// 				"apiVersion": "v1",
// 				"kind":       "Pod",
// 				"metadata": map[string]interface{}{
// 					"name": "test-pod",
// 				},
// 			},
// 			curr: map[string]interface{}{
// 				"apiVersion": "v1",
// 				"kind":       "Pod",
// 				"metadata": map[string]interface{}{
// 					"name":      "test-pod",
// 					"namespace": "default",
// 				},
// 			},
// 			expectedDiff: `
//   metadata:
// +   namespace: default
// `,
// 		},
// 	}

// 	// Iterate over test cases
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// Create unstructured objects for prev and curr
// 			prev := unstructured.Unstructured{Object: tc.prev}
// 			curr := unstructured.Unstructured{Object: tc.curr}

// 			// Create a handler instance
// 			handler, _ := New(HandlerParams{})

// 			// Call the function under test
// 			diff, err := handler.objDiff(prev, curr)

// 			// Assertions
// 			assert.NoError(t, err)

// 			assert.Equal(t, tc.expectedDiff, diff)
// 		})
// 	}
// }

func TestGetKubernetesResourcesFromManifest(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name             string
		manifest         string
		defaultNamespace string
		expectedResults  []struct {
			groupVersionKind schema.GroupVersionKind
			name             string
			namespace        string
		}
	}{
		{
			name: "Manifest with explicit namespaces",
			manifest: `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: test-namespace
spec:
  containers:
  - name: test-container
    image: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: test-namespace
spec:
  replicas: 2
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: test-container
        image: nginx
`,
			defaultNamespace: "",
			expectedResults: []struct {
				groupVersionKind schema.GroupVersionKind
				name             string
				namespace        string
			}{
				{schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}, "test-pod", "test-namespace"},
				{schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, "test-deployment", "test-namespace"},
			},
		},
		{
			name: "Manifest with missing namespace and default namespace applied",
			manifest: `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: test-namespace
spec:
  replicas: 2
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: test-container
        image: nginx
`,
			defaultNamespace: "random",
			expectedResults: []struct {
				groupVersionKind schema.GroupVersionKind
				name             string
				namespace        string
			}{
				{schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}, "test-pod", "random"},
				{schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, "test-deployment", "test-namespace"},
			},
		},
	}

	// Create a fake discovery client
	fakeDiscovery := &discoveryFake.FakeDiscovery{
		Fake: &kubetesting.Fake{},
	}
	fakeDiscovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Kind: "Pod", Namespaced: true},
			},
		},
		{
			GroupVersion: "apps/v1",
			APIResources: []metav1.APIResource{
				{Name: "deployments", Kind: "Deployment", Namespaced: true},
			},
		},
	}

	// Create a discovery mapper
	discoveryMapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(fakeDiscovery))
	// Create a fake discovery client
	fakeDiscoveryClient := &discoveryFake.FakeDiscovery{
		Fake: &kubetesting.Fake{},
	}
	fakeDiscoveryClient.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Kind: "Pod", Namespaced: true},
			},
		},
		{
			GroupVersion: "apps/v1",
			APIResources: []metav1.APIResource{
				{Name: "deployments", Kind: "Deployment", Namespaced: true},
			},
		},
	}
	k := kubernetesClient{
		discoveryMapper: discoveryMapper,
		discoveryClient: fakeDiscoveryClient,
	}

	// Iterate over test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a handler instance
			handler, _ := New(HandlerParams{})
			handler.state = &handlerState{
				plan: &plantypes.DeployPlan{
					KubernetesManifestDeployPlan: &plantypes.KubernetesManifestDeployPlan{
						Namespace: "random",
					},
				},
			}

			resources, err := handler.getKubernetesResourcesFromManifest(&k, tc.manifest)

			// Assertions
			assert.NoError(t, err)
			assert.Len(t, resources, len(tc.expectedResults))

			// Validate each resource
			for i, expected := range tc.expectedResults {
				assert.Equal(t, expected.groupVersionKind, resources[i].groupVersionKind)
				assert.Equal(t, expected.name, resources[i].name)
				assert.Equal(t, expected.namespace, resources[i].namespace)
			}
		})
	}
}
