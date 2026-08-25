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
func TestHydrateFailedPodOwnersReachesDeployment(t *testing.T) {
	e := &Engine{l: zap.NewNop()}

	p := failedPod("api-1-abc", "prod", "ReplicaSet", "api-1", "ImagePullBackOff", "cannot pull")
	rs := controller("ReplicaSet", "api-1", "prod", "Deployment", "api")
	dep := controller("Deployment", "api", "prod", "", "")

	// Only the pod is listed; the owners exist in-cluster but must be fetched.
	byKey := objIndex(p)

	e.hydrateFailedPodOwners(context.Background(), fakeDyn(rs, dep), failedPods(byKey), byKey)

	require.Contains(t, byKey, resourceKey("ReplicaSet", "prod", "api-1"), "replicaset should have been fetched")
	require.Contains(t, byKey, resourceKey("Deployment", "prod", "api"), "deployment should have been fetched")

	assert.Contains(t, liftPodHealthToOwnersIn(byKey), resourceKey("Deployment", "prod", "api"),
		"the failure must land on the deployment, not stop at the replicaset")
}

func TestHydrateFailedPodOwnersSkipsHealthyAndUnowned(t *testing.T) {
	e := &Engine{l: zap.NewNop()}
	dyn := fakeDyn(controller("ReplicaSet", "api-1", "prod", "Deployment", "api"))

	t.Run("healthy pod is not walked", func(t *testing.T) {
		byKey := objIndex(pod("api-1-abc", "prod", "ReplicaSet", "api-1", true))
		e.hydrateFailedPodOwners(context.Background(), dyn, failedPods(byKey), byKey)
		assert.Len(t, byKey, 1, "a healthy pod costs no GETs")
	})

	t.Run("starting pod is not walked", func(t *testing.T) {
		byKey := objIndex(failedPod("p", "prod", "ReplicaSet", "api-1", "ContainerCreating", ""))
		e.hydrateFailedPodOwners(context.Background(), dyn, failedPods(byKey), byKey)
		assert.Len(t, byKey, 1, "a pod still starting up costs no GETs")
	})

	t.Run("missing owner is tolerated", func(t *testing.T) {
		byKey := objIndex(failedPod("orphan-1", "prod", "ReplicaSet", "gone", "ImagePullBackOff", "x"))
		e.hydrateFailedPodOwners(context.Background(), dyn, failedPods(byKey), byKey)
		assert.Len(t, byKey, 1, "a vanished owner must not panic or invent a row")
	})
}
