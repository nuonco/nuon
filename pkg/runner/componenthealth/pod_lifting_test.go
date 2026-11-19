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

// The Always restart policy is what upstream requires before it reads container
// waiting reasons.
func failedPod(name, namespace, ownerKind, ownerName, waitingReason, waitingMessage string) *unstructured.Unstructured {
	u := pod(name, namespace, ownerKind, ownerName, false)
	u.Object["spec"] = map[string]any{"restartPolicy": "Always"}
	status := u.Object["status"].(map[string]any)
	status["phase"] = "Pending"
	status["containerStatuses"] = []any{
		map[string]any{
			"name": "app",
			"state": map[string]any{"waiting": map[string]any{
				"reason":  waitingReason,
				"message": waitingMessage,
			}},
		},
	}
	return u
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

var deploymentGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

// Pods are never attributed to a component, so the controller is the only thing
// that reaches a verdict.
func TestLiftPodHealthToOwnersImagePullBackOff(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "whoami", "whoami", "", "")
	rs := controller("ReplicaSet", "whoami-664778bf79", "whoami", "Deployment", "whoami")
	p := failedPod("whoami-664778bf79-bh5pz", "whoami", "ReplicaSet", "whoami-664778bf79",
		"ImagePullBackOff", `Back-off pulling image "traefik/whoami:does-not-exist"`)

	lifted := liftPodHealthToOwnersIn(objIndex(dep, rs, p))

	got, ok := lifted[resourceKey("Deployment", "whoami", "whoami")]
	require.True(t, ok, "the wedged pod must reach the Deployment across two owner hops")
	assert.Contains(t, got, "Back-off pulling image")

	res := resourceModel(deploymentGVR, dep, nil, got)
	assert.Equal(t, healthDegraded, res.Health,
		"a Deployment whose pod cannot pull its image is not benignly progressing")
	assert.Contains(t, res.Message, "Back-off pulling image")
}

// Keying off status means recovery needs no window to expire.
func TestLiftPodHealthToOwnersIgnoresRecoveredPod(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "api", "prod", "", "")
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	p := pod("api-1-abc", "prod", "ReplicaSet", "api-1", true)

	lifted := liftPodHealthToOwnersIn(objIndex(dep, rs, p))

	assert.NotContains(t, lifted, resourceKey("Deployment", "prod", "api"),
		"a pod that recovered must not hold its controller degraded at all")
}

// Lifting a starting pod would degrade every rollout.
func TestLiftPodHealthToOwnersIgnoresStartingPod(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "api", "prod", "", "")
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	p := failedPod("api-1-abc", "prod", "ReplicaSet", "api-1", "ContainerCreating", "")

	lifted := liftPodHealthToOwnersIn(objIndex(dep, rs, p))

	assert.Empty(t, lifted, "ContainerCreating is not a failure")
}

// Otherwise the reported message churns with map iteration order.
func TestLiftPodHealthToOwnersIsDeterministic(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "api", "prod", "", "")
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	a := failedPod("api-1-aaa", "prod", "ReplicaSet", "api-1", "ImagePullBackOff", "first pod")
	b := failedPod("api-1-bbb", "prod", "ReplicaSet", "api-1", "CrashLoopBackOff", "second pod")

	key := resourceKey("Deployment", "prod", "api")
	for i := 0; i < 20; i++ {
		lifted := liftPodHealthToOwnersIn(objIndex(dep, rs, a, b))
		assert.Equal(t, "first pod", lifted[key])
	}
}

func TestLiftPodHealthToOwnersOrphanPod(t *testing.T) {
	t.Parallel()

	p := failedPod("standalone", "prod", "", "", "ImagePullBackOff", "no owner")

	assert.Empty(t, liftPodHealthToOwnersIn(objIndex(p)),
		"a pod with no controller owner lifts nothing and must not panic")
}

// A controller's own failure is more specific than anything lifted from a child.
func TestLiftedFailureDoesNotOverrideOwnStatus(t *testing.T) {
	t.Parallel()

	dep := deploymentWithConditions(2,
		hpaCond("Available", "True", "MinimumReplicasAvailable", "Deployment has minimum availability."),
		hpaCond("ReplicaFailure", "True", "FailedCreate", "exceeded quota"),
	)

	res := resourceModel(deploymentGVR, dep, nil, "some pod is unhappy")

	assert.Equal(t, healthDegraded, res.Health)
	assert.Contains(t, res.Message, "exceeded quota")
}

func TestTopOwnerStopsOnOwnerNotListed(t *testing.T) {
	t.Parallel()

	// The ReplicaSet is absent from the listing (e.g. the list call for
	// replicasets failed), so the walk must stop rather than guess.
	p := pod("api-1-abc", "prod", "ReplicaSet", "api-1", false)

	assert.Nil(t, topOwner(p, objIndex(p)))
}

// An evicted pod lingers until GC, long after its replacement came up.
func TestLiftPodHealthToOwnersIgnoresTerminatedPod(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "api", "prod", "", "")
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	evicted := failedPod("api-1-old", "prod", "ReplicaSet", "api-1", "Evicted", "the node was low on resource: memory")
	evicted.Object["status"].(map[string]any)["phase"] = "Failed"

	assert.Empty(t, liftPodHealthToOwnersIn(objIndex(dep, rs, evicted)),
		"a replaced pod must not hold its controller degraded until garbage collection")
}
