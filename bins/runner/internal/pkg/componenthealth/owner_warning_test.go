package componenthealth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func pod(name, namespace, ownerKind, ownerName string, ready bool) *unstructured.Unstructured {
	readyStatus := "False"
	if ready {
		readyStatus = "True"
	}
	obj := map[string]any{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
		},
		"status": map[string]any{
			"conditions": []any{
				map[string]any{"type": "Ready", "status": readyStatus},
			},
		},
	}
	if ownerKind != "" {
		obj["metadata"].(map[string]any)["ownerReferences"] = []any{
			map[string]any{
				"apiVersion": "apps/v1",
				"kind":       ownerKind,
				"name":       ownerName,
				"controller": true,
				"uid":        "uid-" + ownerName,
			},
		}
	}
	return &unstructured.Unstructured{Object: obj}
}

func controller(kind, name, namespace, ownerKind, ownerName string) *unstructured.Unstructured {
	obj := map[string]any{
		"apiVersion": "apps/v1",
		"kind":       kind,
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
		},
	}
	if ownerKind != "" {
		obj["metadata"].(map[string]any)["ownerReferences"] = []any{
			map[string]any{
				"apiVersion": "apps/v1",
				"kind":       ownerKind,
				"name":       ownerName,
				"controller": true,
				"uid":        "uid-" + ownerName,
			},
		}
	}
	return &unstructured.Unstructured{Object: obj}
}

func objIndex(objs ...*unstructured.Unstructured) map[string]*unstructured.Unstructured {
	out := map[string]*unstructured.Unstructured{}
	for _, o := range objs {
		out[resourceKey(o.GetKind(), o.GetNamespace(), o.GetName())] = o
	}
	return out
}

func TestLiftPodWarningsToOwnersImagePullBackOff(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "whoami", "whoami", "", "")
	rs := controller("ReplicaSet", "whoami-664778bf79", "whoami", "Deployment", "whoami")
	p := pod("whoami-664778bf79-bh5pz", "whoami", "ReplicaSet", "whoami-664778bf79", false)

	warnings := map[string]warningEvent{
		resourceKey("Pod", "whoami", "whoami-664778bf79-bh5pz"): {
			reason:  "Failed",
			message: `Failed to pull image "traefik/whoami:does-not-exist": not found`,
		},
	}

	liftPodWarningsToOwners(warnings, objIndex(dep, rs, p))

	got, ok := warnings[resourceKey("Deployment", "whoami", "whoami")]
	require.True(t, ok, "the wedged pod's warning must reach the Deployment across two owner hops")
	assert.Equal(t, "Failed", got.reason)

	// resourceModel is what turns that warning into a verdict; assert the
	// end-to-end consequence, since "progressing" is what the bug produced.
	res := resourceModel(schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, dep, &got)
	assert.Equal(t, healthDegraded, res.Health,
		"a Deployment whose pod cannot pull its image is not benignly progressing")
	assert.Contains(t, res.Message, "Failed to pull image")
}

func TestLiftPodWarningsToOwnersIgnoresReadyPod(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "api", "prod", "", "")
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	p := pod("api-1-abc", "prod", "ReplicaSet", "api-1", true)

	warnings := map[string]warningEvent{
		resourceKey("Pod", "prod", "api-1-abc"): {reason: "FailedScheduling", message: "transient, since recovered"},
	}

	liftPodWarningsToOwners(warnings, objIndex(dep, rs, p))

	_, ok := warnings[resourceKey("Deployment", "prod", "api")]
	assert.False(t, ok,
		"a pod that recovered must not hold its controller degraded for the whole event window")
}

func TestLiftPodWarningsToOwnersKeepsControllerOwnWarning(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "api", "prod", "", "")
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	p := pod("api-1-abc", "prod", "ReplicaSet", "api-1", false)

	warnings := map[string]warningEvent{
		resourceKey("Deployment", "prod", "api"): {reason: "ReplicaSetCreateError", message: "specific to the controller"},
		resourceKey("Pod", "prod", "api-1-abc"):  {reason: "Failed", message: "less specific"},
	}

	liftPodWarningsToOwners(warnings, objIndex(dep, rs, p))

	assert.Equal(t, "ReplicaSetCreateError", warnings[resourceKey("Deployment", "prod", "api")].reason,
		"a warning on the controller itself is more specific and must win")
}

func TestLiftPodWarningsToOwnersOrphanPod(t *testing.T) {
	t.Parallel()

	p := pod("standalone", "prod", "", "", false)
	warnings := map[string]warningEvent{
		resourceKey("Pod", "prod", "standalone"): {reason: "Failed", message: "no owner"},
	}

	liftPodWarningsToOwners(warnings, objIndex(p))

	assert.Len(t, warnings, 1, "a pod with no controller owner lifts nothing and must not panic")
}

func TestTopOwnerStopsOnOwnerNotListed(t *testing.T) {
	t.Parallel()

	// The ReplicaSet is absent from the listing (e.g. the list call for
	// replicasets failed), so the walk must stop rather than guess.
	p := pod("api-1-abc", "prod", "ReplicaSet", "api-1", false)

	assert.Nil(t, topOwner(p, objIndex(p)))
}

func TestPodReadyMissingConditions(t *testing.T) {
	t.Parallel()

	u := &unstructured.Unstructured{Object: map[string]any{
		"kind":     "Pod",
		"metadata": map[string]any{"name": "p", "namespace": "n"},
	}}

	assert.False(t, podReady(u), "a pod with no Ready condition is not ready")
}
