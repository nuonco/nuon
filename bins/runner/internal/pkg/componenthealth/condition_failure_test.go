package componenthealth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var hpaGVR = schema.GroupVersionResource{Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers"}

func hpaCond(condType, status, reason, message string) map[string]any {
	return map[string]any{"type": condType, "status": status, "reason": reason, "message": message}
}

func hpaObj(name string, conds ...map[string]any) *unstructured.Unstructured {
	raw := make([]any, 0, len(conds))
	for _, c := range conds {
		raw = append(raw, c)
	}
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "autoscaling/v2",
		"kind":       "HorizontalPodAutoscaler",
		"metadata":   map[string]any{"name": name, "namespace": "kitchen-sink"},
		"spec":       map[string]any{"minReplicas": int64(1), "maxReplicas": int64(3)},
		"status":     map[string]any{"conditions": raw},
	}}
}

// Kubernetes orders AbleToScale ahead of ScalingActive, and upstream returns on
// the first condition matched.
func TestHPAMetricFailureIsReadFromStatus(t *testing.T) {
	t.Parallel()

	obj := hpaObj("kitchen-sink-api",
		hpaCond("AbleToScale", "True", "ReadyForNewScale", "recommended size matches current size"),
		hpaCond("ScalingActive", "False", "FailedGetResourceMetric", "failed to get cpu utilization: did not receive metrics for targeted pods"),
	)

	health, message, _ := assessResource(obj)
	assert.Equal(t, healthDegraded, health, "AbleToScale=True must not mask ScalingActive=False")
	assert.Contains(t, message, "did not receive metrics")
}

func TestHPAHealthyWhenScalingActive(t *testing.T) {
	t.Parallel()

	obj := hpaObj("kitchen-sink-ui",
		hpaCond("AbleToScale", "True", "ReadyForNewScale", "recommended size matches current size"),
		hpaCond("ScalingActive", "True", "ValidMetricFound", "the HPA was able to compute the replica count"),
	)

	health, _, _ := assessResource(obj)
	assert.Equal(t, healthHealthy, health)
}

// A target scaled to zero, not a failure.
func TestHPAScalingDisabledIsNotAFailure(t *testing.T) {
	t.Parallel()

	obj := hpaObj("worker",
		hpaCond("AbleToScale", "True", "SucceededGetScale", "the HPA controller was able to get the target's current scale"),
		hpaCond("ScalingActive", "False", "ScalingDisabled", "scaling is disabled since the replica count of the target is zero"),
	)

	health, _, _ := assessResource(obj)
	assert.Equal(t, healthHealthy, health)
}

func deploymentWithConditions(replicas int64, conds ...map[string]any) *unstructured.Unstructured {
	raw := make([]any, 0, len(conds))
	for _, c := range conds {
		raw = append(raw, c)
	}
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata":   map[string]any{"name": "api", "namespace": "prod", "generation": int64(1)},
		"spec":       map[string]any{"replicas": replicas},
		"status": map[string]any{
			"observedGeneration": int64(1),
			"replicas":           replicas,
			"updatedReplicas":    replicas,
			"availableReplicas":  replicas,
			"readyReplicas":      replicas,
			"conditions":         raw,
		},
	}}
}

// Not an HPA problem: upstream reads only Progressing here, so a Deployment that
// cannot create pods reads healthy whenever replica counts line up.
func TestDeploymentReplicaFailureIsRead(t *testing.T) {
	t.Parallel()

	obj := deploymentWithConditions(2,
		hpaCond("Available", "True", "MinimumReplicasAvailable", "Deployment has minimum availability."),
		hpaCond("ReplicaFailure", "True", "FailedCreate", `pods "api-" is forbidden: exceeded quota`),
	)

	health, message, _ := assessResource(obj)
	assert.Equal(t, healthDegraded, health, "a Deployment that cannot create pods is not healthy")
	assert.Contains(t, message, "exceeded quota")
}

func TestHealthyDeploymentStaysHealthy(t *testing.T) {
	t.Parallel()

	obj := deploymentWithConditions(2,
		hpaCond("Available", "True", "MinimumReplicasAvailable", "Deployment has minimum availability."),
		hpaCond("Progressing", "True", "NewReplicaSetAvailable", "ReplicaSet has successfully progressed."),
	)

	health, _, _ := assessResource(obj)
	assert.Equal(t, healthHealthy, health)
}

// A reason left behind on a now-True ready condition is already over.
func TestRecoveredReadyConditionIsNotAFailure(t *testing.T) {
	t.Parallel()

	obj := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "cert-manager.io/v1",
		"kind":       "Certificate",
		"metadata":   map[string]any{"name": "cert", "namespace": "prod"},
		"status": map[string]any{"conditions": []any{
			hpaCond("Ready", "True", "ErrGetKeyPair", "issued"),
		}},
	}}

	health, _, _ := assessResource(obj)
	assert.Equal(t, healthHealthy, health)
}

func TestFailureReasonNaming(t *testing.T) {
	t.Parallel()

	for _, reason := range []string{"FailedGetResourceMetric", "FailedCreate", "ErrImagePull", "InvalidSelector", "UpgradeFailed", "SyncError", "CrashLoopBackOff"} {
		assert.True(t, failureReason(reason), reason)
	}
	for _, reason := range []string{"", "ReadyForNewScale", "ScalingDisabled", "MinimumReplicasAvailable", "NewReplicaSetAvailable", "TooManyReplicas"} {
		assert.False(t, failureReason(reason), reason)
	}
}

// The API conventions define a condition set against an older generation as out
// of date, so it must not produce a verdict either way.
func TestStaleConditionIsIgnored(t *testing.T) {
	t.Parallel()

	obj := deploymentWithConditions(2,
		hpaCond("Available", "True", "MinimumReplicasAvailable", "Deployment has minimum availability."),
		hpaCond("ReplicaFailure", "True", "FailedCreate", "exceeded quota"),
	)
	obj.SetGeneration(9)
	// Object-level generation must match, or the whole object reads progressing
	// and the per-condition rule is never reached.
	_ = unstructured.SetNestedField(obj.Object, int64(9), "status", "observedGeneration")
	conds, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	conds[1].(map[string]any)["observedGeneration"] = int64(8)
	_ = unstructured.SetNestedSlice(obj.Object, conds, "status", "conditions")

	health, _, _ := assessResource(obj)
	assert.Equal(t, healthHealthy, health, "a failure from generation 8 says nothing about generation 9")

	conds, _, _ = unstructured.NestedSlice(obj.Object, "status", "conditions")
	conds[1].(map[string]any)["observedGeneration"] = int64(9)
	_ = unstructured.SetNestedSlice(obj.Object, conds, "status", "conditions")

	health, _, _ = assessResource(obj)
	assert.Equal(t, healthDegraded, health)
}

// A condition without the field is the common case and must stay trusted.
func TestConditionWithoutObservedGenerationIsTrusted(t *testing.T) {
	t.Parallel()

	assert.False(t, staleCondition(map[string]any{"reason": "FailedCreate"}, 12))
	assert.False(t, staleCondition(map[string]any{"observedGeneration": int64(0)}, 12))
	assert.True(t, staleCondition(map[string]any{"observedGeneration": float64(9)}, 12))
}
