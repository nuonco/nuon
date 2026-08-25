package componenthealth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func fakeDyn(objs ...runtime.Object) *dynamicfake.FakeDynamicClient {
	scheme := runtime.NewScheme()
	gvrToKind := map[schema.GroupVersionResource]string{
		{Group: "apps", Version: "v1", Resource: "replicasets"}:  "ReplicaSetList",
		{Group: "apps", Version: "v1", Resource: "deployments"}:  "DeploymentList",
		{Group: "apps", Version: "v1", Resource: "statefulsets"}: "StatefulSetList",
		{Group: "apps", Version: "v1", Resource: "daemonsets"}:   "DaemonSetList",
		{Group: "batch", Version: "v1", Resource: "jobs"}:        "JobList",
	}
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, gvrToKind, objs...)
}

// The regression this guards: ReplicaSets are no longer listed, so an
// ImagePullBackOff on a pod must still reach its Deployment via on-demand GETs.
// Without hydration the walk stops at the unlisted ReplicaSet and the rollout
// reads as benign "progressing" for progressDeadlineSeconds (10m).
func TestHydrateWarningOwnersReachesDeployment(t *testing.T) {
	e := &Engine{l: zap.NewNop()}

	p := pod("api-1-abc", "prod", "ReplicaSet", "api-1", false)
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	dep := controller("Deployment", "api", "prod", "", "")

	// Only the pod is listed; the owners exist in-cluster but must be fetched.
	byKey := objIndex(p)
	warn := warningFor(p, "Failed", "ImagePullBackOff")
	warnings := map[string]warningEvent{
		warn.source.key(): warn,
	}

	e.hydrateWarningOwners(context.Background(), fakeDyn(rs, dep), warnings, byKey)

	require.Contains(t, byKey, resourceRefForObject(rs).key(), "replicaset should have been fetched")
	require.Contains(t, byKey, resourceRefForObject(dep).key(), "deployment should have been fetched")

	attributeWarningsToOwners(warnings, byKey)
	assert.Contains(t, warnings, resourceRefForObject(dep).key(),
		"the warning must land on the deployment, not stop at the replicaset")
}

func TestHydrateWarningOwnersSkipsHealthyAndUnowned(t *testing.T) {
	e := &Engine{l: zap.NewNop()}
	dyn := fakeDyn(controller("ReplicaSet", "api-1", "prod", "Deployment", "api"))

	t.Run("ready pod is not walked", func(t *testing.T) {
		p := pod("api-1-abc", "prod", "ReplicaSet", "api-1", true)
		byKey := objIndex(p)
		warn := warningFor(p, "X", "")
		e.hydrateWarningOwners(context.Background(), dyn,
			map[string]warningEvent{warn.source.key(): warn}, byKey)
		assert.Len(t, byKey, 1, "a ready pod costs no GETs")
	})

	t.Run("no warnings means no work", func(t *testing.T) {
		byKey := objIndex(pod("p", "prod", "ReplicaSet", "api-1", false))
		e.hydrateWarningOwners(context.Background(), dyn, map[string]warningEvent{}, byKey)
		assert.Len(t, byKey, 1)
	})

	t.Run("missing watched source is not fetched", func(t *testing.T) {
		missing := pod("gone", "prod", "ReplicaSet", "api-1", false)
		warn := warningFor(missing, "Failed", "stale event")
		client := fakeDyn(missing)
		e.hydrateWarningOwners(context.Background(), client,
			map[string]warningEvent{warn.source.key(): warn}, objIndex())
		assert.Empty(t, client.Actions())
	})

	t.Run("missing owner is tolerated", func(t *testing.T) {
		p := pod("orphan-1", "prod", "ReplicaSet", "gone", false)
		byKey := objIndex(p)
		warn := warningFor(p, "X", "")
		e.hydrateWarningOwners(context.Background(), dyn,
			map[string]warningEvent{warn.source.key(): warn}, byKey)
		assert.Len(t, byKey, 1, "a vanished owner must not panic or invent a row")
	})
}

func TestHydrateFailedCreateSourceWithoutPod(t *testing.T) {
	e := &Engine{l: zap.NewNop()}
	dep := controller("Deployment", "api", "prod", "", "")
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	warn := warningFor(rs, "FailedCreate", "admission webhook denied the request")
	warnings := map[string]warningEvent{warn.source.key(): warn}
	byKey := objIndex(dep)

	e.hydrateWarningOwners(context.Background(), fakeDyn(rs), warnings, byKey)
	attributeWarningsToOwners(warnings, byKey)

	assert.Contains(t, byKey, resourceRefForObject(rs).key())
	assert.Contains(t, warnings, resourceRefForObject(dep).key())
}
