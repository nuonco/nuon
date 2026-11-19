package componenthealth

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// testdata/live_objects.json is captured verbatim from a real cluster running
// the failure modes below, so hand-written fixtures cannot drift from what
// Kubernetes actually writes.
func loadLiveObjects(t *testing.T) map[string]*unstructured.Unstructured {
	t.Helper()

	raw, err := os.ReadFile("testdata/live_objects.json")
	require.NoError(t, err)

	var objs map[string]map[string]any
	require.NoError(t, json.Unmarshal(raw, &objs))

	out := make(map[string]*unstructured.Unstructured, len(objs))
	for key, obj := range objs {
		out[key] = &unstructured.Unstructured{Object: obj}
	}
	return out
}

func TestLiveObjectVerdicts(t *testing.T) {
	t.Parallel()

	byKey := loadLiveObjects(t)
	lifted := liftPodHealthToOwners(failedPods(byKey), byKey)

	want := map[string]string{
		"Deployment/health-demo/healthy-api":              healthHealthy,
		"Pod/health-demo/healthy-api-6799649d5f-97trm":    healthHealthy,
		"HorizontalPodAutoscaler/health-demo/healthy-api": healthDegraded,
		"Deployment/health-quota/quota-blocked":           healthDegraded,
		"Pod/health-demo/bad-image-7b84f6f4f6-cnrj9":      healthDegraded,
		"Pod/health-demo/crashloop-56759cc757-wskzd":      healthDegraded,
		"Deployment/health-demo/bad-image":                healthDegraded,
		"Deployment/health-demo/crashloop":                healthDegraded,
		"ReplicaSet/health-demo/bad-image-7b84f6f4f6":     healthProgressing,
		"ReplicaSet/health-demo/crashloop-56759cc757":     healthProgressing,
	}

	for key, expected := range want {
		obj, ok := byKey[key]
		require.True(t, ok, key)
		got := resourceModel(schema.GroupVersionResource{}, obj, nil, lifted[key])
		assert.Equal(t, expected, got.Health, key)
		if expected != healthHealthy {
			assert.NotEmpty(t, got.Message, "%s: a non-healthy verdict must say why", key)
		}
	}
}

// Every verdict above must hold with an arbitrary event attached.
func TestLiveObjectVerdictsIgnoreEvents(t *testing.T) {
	t.Parallel()

	byKey := loadLiveObjects(t)
	lifted := liftPodHealthToOwners(failedPods(byKey), byKey)
	warn := &warningEvent{reason: "FailedSomething", message: "fired minutes ago"}

	for key, obj := range byKey {
		base := resourceModel(schema.GroupVersionResource{}, obj, nil, lifted[key])
		withEvent := resourceModel(schema.GroupVersionResource{}, obj, warn, lifted[key])
		assert.Equal(t, base.Health, withEvent.Health, key)
	}
}

// The incident: upstream grades this HPA healthy because Kubernetes writes
// AbleToScale first, which left the metrics failure reachable only as an event.
func TestLiveHPAFailureCarriesTheRealReason(t *testing.T) {
	t.Parallel()

	obj := loadLiveObjects(t)["HorizontalPodAutoscaler/health-demo/healthy-api"]
	require.NotNil(t, obj)

	health, message, native := assessResource(obj)
	assert.Equal(t, healthDegraded, health)
	assert.Contains(t, message, "failed to get cpu utilization")
	assert.Equal(t, "ScalingActive=False/FailedGetResourceMetric", native)
}

// A Deployment blocked by quota reports availableReplicas matching its own
// expectations, so only ReplicaFailure reveals it.
func TestLiveQuotaBlockedDeployment(t *testing.T) {
	t.Parallel()

	obj := loadLiveObjects(t)["Deployment/health-quota/quota-blocked"]
	require.NotNil(t, obj)

	health, message, _ := assessResource(obj)
	assert.Equal(t, healthDegraded, health)
	assert.Contains(t, message, "exceeded quota")
}

// Lifting crosses two owner hops, and the pod's own status.message is empty
// here so the reason has to come from the container status.
func TestLiveCrashLoopReachesDeployment(t *testing.T) {
	t.Parallel()

	byKey := loadLiveObjects(t)
	lifted := liftPodHealthToOwners(failedPods(byKey), byKey)

	assert.Contains(t, lifted["Deployment/health-demo/crashloop"], "container app")
	assert.Contains(t, lifted["Deployment/health-demo/bad-image"], "Back-off pulling image")
}
