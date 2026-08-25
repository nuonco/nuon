package componenthealth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestEventObjectRefPreservesKubernetesIdentity(t *testing.T) {
	event := &unstructured.Unstructured{Object: map[string]any{
		"involvedObject": map[string]any{
			"apiVersion": "apps/v1",
			"kind":       "ReplicaSet",
			"namespace":  "prod",
			"name":       "api-1",
			"uid":        "uid-api-1",
		},
	}}

	ref, ok := eventObjectRef(event)
	require.True(t, ok)
	assert.Equal(t, resourceRef{
		APIVersion: "apps/v1",
		Kind:       "ReplicaSet",
		Namespace:  "prod",
		Name:       "api-1",
		UID:        "uid-api-1",
	}, ref)
}

func TestFallbackResourceMappingUsesFullGVK(t *testing.T) {
	gvr, ok := resourceGVR(resourceRef{APIVersion: "apps/v1", Kind: "ReplicaSet"})
	require.True(t, ok)
	assert.Equal(t, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicasets"}, gvr)

	_, ok = resourceGVR(resourceRef{APIVersion: "example.io/v1", Kind: "ReplicaSet"})
	assert.False(t, ok)
}
