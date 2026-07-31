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
	warnings := map[string]warningEvent{
		resourceKey("Pod", "prod", "api-1-abc"): {reason: "Failed", message: "ImagePullBackOff"},
	}

	e.hydrateWarningOwners(context.Background(), fakeDyn(rs, dep), warnings, byKey)

	require.Contains(t, byKey, resourceKey("ReplicaSet", "prod", "api-1"), "replicaset should have been fetched")
	require.Contains(t, byKey, resourceKey("Deployment", "prod", "api"), "deployment should have been fetched")

	liftPodWarningsToOwners(warnings, byKey)
	assert.Contains(t, warnings, resourceKey("Deployment", "prod", "api"),
		"the warning must land on the deployment, not stop at the replicaset")
}

func TestHydrateWarningOwnersSkipsHealthyAndUnowned(t *testing.T) {
	e := &Engine{l: zap.NewNop()}
	dyn := fakeDyn(controller("ReplicaSet", "api-1", "prod", "Deployment", "api"))

	t.Run("ready pod is not walked", func(t *testing.T) {
		p := pod("api-1-abc", "prod", "ReplicaSet", "api-1", true)
		byKey := objIndex(p)
		e.hydrateWarningOwners(context.Background(), dyn,
			map[string]warningEvent{resourceKey("Pod", "prod", "api-1-abc"): {reason: "X"}}, byKey)
		assert.Len(t, byKey, 1, "a ready pod costs no GETs")
	})

	t.Run("no warnings means no work", func(t *testing.T) {
		byKey := objIndex(pod("p", "prod", "ReplicaSet", "api-1", false))
		e.hydrateWarningOwners(context.Background(), dyn, map[string]warningEvent{}, byKey)
		assert.Len(t, byKey, 1)
	})

	t.Run("missing owner is tolerated", func(t *testing.T) {
		p := pod("orphan-1", "prod", "ReplicaSet", "gone", false)
		byKey := objIndex(p)
		e.hydrateWarningOwners(context.Background(), dyn,
			map[string]warningEvent{resourceKey("Pod", "prod", "orphan-1"): {reason: "X"}}, byKey)
		assert.Len(t, byKey, 1, "a vanished owner must not panic or invent a row")
	})
}
