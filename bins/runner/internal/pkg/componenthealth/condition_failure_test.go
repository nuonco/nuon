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

// The bug that started this: Kubernetes orders AbleToScale ahead of
// ScalingActive and the upstream check returns on the first condition that
// matches anything, so a metrics failure was unreadable from status and only
// reachable through a Warning event.
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

// ScalingDisabled is what a target scaled to zero looks like, not a failure.
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

// Not an HPA problem: the upstream Deployment check consults only the
// Progressing condition, so a Deployment that cannot create pods reads healthy
// whenever the replica counts happen to line up.
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

// A failure reason left behind on a ready-style condition that now reads True
// describes something already over.
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
