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
			"uid":       "uid-" + name,
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
			"uid":       "uid-" + name,
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
		out[resourceRefForObject(o).key()] = o
	}
	return out
}

func warningFor(u *unstructured.Unstructured, reason, message string) warningEvent {
	return warningEvent{reason: reason, message: message, source: resourceRefForObject(u)}
}

func TestLiftPodWarningsToOwnersImagePullBackOff(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "whoami", "whoami", "", "")
	rs := controller("ReplicaSet", "whoami-664778bf79", "whoami", "Deployment", "whoami")
	p := pod("whoami-664778bf79-bh5pz", "whoami", "ReplicaSet", "whoami-664778bf79", false)

	warnings := map[string]warningEvent{
		resourceRefForObject(p).key(): warningFor(
			p,
			"Failed",
			`Failed to pull image "traefik/whoami:does-not-exist": not found`,
		),
	}

	attributeWarningsToOwners(warnings, objIndex(dep, rs, p))

	got, ok := warnings[resourceRefForObject(dep).key()]
	require.True(t, ok, "the wedged pod's warning must reach the Deployment across two owner hops")
	assert.Equal(t, "Failed", got.reason)
	assert.Equal(t, []resourceRef{
		resourceRefForObject(p),
		resourceRefForObject(rs),
		resourceRefForObject(dep),
	}, got.ownerPath)

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
		resourceRefForObject(p).key(): warningFor(p, "FailedScheduling", "transient, since recovered"),
	}

	attributeWarningsToOwners(warnings, objIndex(dep, rs, p))

	_, ok := warnings[resourceRefForObject(dep).key()]
	assert.False(t, ok,
		"a pod that recovered must not hold its controller degraded for the whole event window")
}

func TestLiftPodWarningsToOwnersKeepsControllerOwnWarning(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "api", "prod", "", "")
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	p := pod("api-1-abc", "prod", "ReplicaSet", "api-1", false)

	warnings := map[string]warningEvent{
		resourceRefForObject(dep).key(): warningFor(dep, "ReplicaSetCreateError", "specific to the controller"),
		resourceRefForObject(p).key():   warningFor(p, "Failed", "less specific"),
	}

	attributeWarningsToOwners(warnings, objIndex(dep, rs, p))

	assert.Equal(t, "ReplicaSetCreateError", warnings[resourceRefForObject(dep).key()].reason,
		"a warning on the controller itself is more specific and must win")
}

func TestLiftPodWarningsToOwnersReplacesStaleControllerWarning(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "api", "prod", "", "")
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	p := pod("api-1-abc", "prod", "ReplicaSet", "api-1", false)
	stale := warningFor(dep, "OldFailure", "from the previous deployment")
	stale.source.UID = "uid-deleted-deployment"

	warnings := map[string]warningEvent{
		resourceRefForObject(dep).key(): stale,
		resourceRefForObject(p).key():   warningFor(p, "Failed", "current failure"),
	}

	attributeWarningsToOwners(warnings, objIndex(dep, rs, p))

	assert.Equal(t, "Failed", warnings[resourceRefForObject(dep).key()].reason)
}

func TestLiftPodWarningsToOwnersOrphanPod(t *testing.T) {
	t.Parallel()

	p := pod("standalone", "prod", "", "", false)
	warnings := map[string]warningEvent{
		resourceRefForObject(p).key(): warningFor(p, "Failed", "no owner"),
	}

	attributeWarningsToOwners(warnings, objIndex(p))

	assert.Len(t, warnings, 1, "a pod with no controller owner lifts nothing and must not panic")
}

func TestAttributeFailedCreateFromReplicaSetToDeployment(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "api", "prod", "", "")
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	warn := warningFor(rs, "FailedCreate", "admission webhook denied the request: signature verification failed")
	warnings := map[string]warningEvent{warn.source.key(): warn}

	attributeWarningsToOwners(warnings, objIndex(dep, rs))

	got, ok := warnings[resourceRefForObject(dep).key()]
	require.True(t, ok)
	assert.Equal(t, "FailedCreate", got.reason)
	res := resourceModel(schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, dep, &got)
	assert.Equal(t, healthDegraded, res.Health)
	assert.Contains(t, res.Details, `"kind":"ReplicaSet"`)
	assert.Contains(t, res.Details, `"owner_path"`)
}

func TestAttributeWarningRejectsRecreatedSource(t *testing.T) {
	t.Parallel()

	dep := controller("Deployment", "api", "prod", "", "")
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	warn := warningFor(rs, "FailedCreate", "old admission failure")
	warn.source.UID = "uid-deleted-replicaset"
	warnings := map[string]warningEvent{warn.source.key(): warn}

	attributeWarningsToOwners(warnings, objIndex(dep, rs))

	assert.NotContains(t, warnings, resourceRefForObject(dep).key(),
		"an event for a deleted object must not degrade a replacement with the same name")
}

func TestOwnerPathStopsOnOwnerNotListed(t *testing.T) {
	t.Parallel()

	// The ReplicaSet is absent from the listing (e.g. the list call for
	// replicasets failed), so the walk must stop rather than guess.
	p := pod("api-1-abc", "prod", "ReplicaSet", "api-1", false)

	assert.Equal(t, []resourceRef{resourceRefForObject(p)}, ownerPath(p, objIndex(p)))
}

func TestPodReadyMissingConditions(t *testing.T) {
	t.Parallel()

	u := &unstructured.Unstructured{Object: map[string]any{
		"kind":     "Pod",
		"metadata": map[string]any{"name": "p", "namespace": "n"},
	}}

	assert.False(t, podReady(u), "a pod with no Ready condition is not ready")
}
