package componenthealth

import (
	"encoding/json"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/componenthealth/gitopshealth"
)

const (
	healthHealthy       = "healthy"
	healthProgressing   = "progressing"
	healthDegraded      = "degraded"
	healthUnhealthy     = "unhealthy"
	healthUnknown       = "unknown"
	healthNotApplicable = "not-applicable"

	maxDetailsBytes = 8 * 1024
)

func assessResource(obj *unstructured.Unstructured) (health, message, nativeStatus string) {
	hs, err := gitopshealth.GetResourceHealth(obj, nil)
	if err != nil {
		return healthUnknown, "", ""
	}
	if hs == nil {
		return assessByConditions(obj)
	}

	health = mapHealth(hs.Status)
	if health == healthHealthy || health == healthProgressing {
		if reason, msg, ok := initContainerFailure(obj); ok {
			return healthDegraded, msg, reason
		}
		if !staleGeneration(obj) {
			if reason, msg, ok := conditionFailure(obj); ok {
				return healthDegraded, msg, reason
			}
		}
	}

	msg := hs.Message
	if msg == "" {
		msg = explainVerdict(obj, hs.Status)
	}
	return health, msg, string(hs.Status)
}

// why: initContainerFailure scans init containers, which the upstream pod check
// skips: it reads containerStatuses only, so an init container stuck on
// ImagePullBackOff left the pod merely Pending with the reason sitting in status.
func initContainerFailure(obj *unstructured.Unstructured) (reason, message string, found bool) {
	if obj.GetKind() != "Pod" {
		return "", "", false
	}
	statuses, ok, err := unstructured.NestedSlice(obj.Object, "status", "initContainerStatuses")
	if err != nil || !ok {
		return "", "", false
	}
	for _, raw := range statuses {
		cs, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		waitReason, _, _ := unstructured.NestedString(cs, "state", "waiting", "reason")
		if !failureReason(waitReason) {
			continue
		}
		name, _ := cs["name"].(string)
		message, _, _ := unstructured.NestedString(cs, "state", "waiting", "message")
		if message == "" {
			message = "init container " + name + ": " + waitReason
		}
		return "init/" + waitReason, message, true
	}
	return "", "", false
}

// why: conditionFailure reads every condition, because each upstream per-kind check
// reads only a slice of status: the HPA check returns on the first condition
// matched, the Deployment check never looks at ReplicaFailure.
func conditionFailure(obj *unstructured.Unstructured) (reason, message string, found bool) {
	conds, ok, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil || !ok {
		return "", "", false
	}

	gen := obj.GetGeneration()
	for _, c := range conds {
		cond, ok := c.(map[string]any)
		if !ok {
			continue
		}
		condReason, _ := cond["reason"].(string)
		if !failureReason(condReason) || staleCondition(cond, gen) {
			continue
		}
		condType, _ := cond["type"].(string)
		condStatus, _ := cond["status"].(string)
		if condStatus == "True" && isReadyConditionType(condType) {
			continue
		}

		condMessage, _ := cond["message"].(string)
		if condMessage == "" {
			condMessage = condReason
		}
		return condType + "=" + condStatus + "/" + condReason, condMessage, true
	}
	return "", "", false
}

// why: Polarity cannot come from status: ReplicaFailure=True and ScalingActive=False
// both mean broken, so the reason is the only consistent signal.
func failureReason(reason string) bool {
	switch {
	case reason == "":
		return false
	case reason == "Unschedulable":
		return true
	case strings.HasPrefix(reason, "Failed"),
		strings.HasPrefix(reason, "Err"),
		strings.HasPrefix(reason, "Invalid"),
		strings.HasSuffix(reason, "Failed"),
		strings.HasSuffix(reason, "Error"),
		strings.HasSuffix(reason, "BackOff"):
		return true
	}
	return false
}

func explainVerdict(obj *unstructured.Unstructured, status gitopshealth.HealthStatusCode) string {
	if obj.GetKind() == "Pod" && status != gitopshealth.HealthStatusHealthy {
		return podFailureReason(obj)
	}
	if status != gitopshealth.HealthStatusProgressing {
		return ""
	}

	switch obj.GetKind() {
	case "Ingress":
		if !hasLoadBalancerAddress(obj) {
			return "no load balancer address assigned yet"
		}
	case "Service":
		if !hasLoadBalancerAddress(obj) {
			return "waiting for a load balancer address"
		}
	}
	return ""
}

func hasLoadBalancerAddress(obj *unstructured.Unstructured) bool {
	addrs, found, err := unstructured.NestedSlice(obj.Object, "status", "loadBalancer", "ingress")
	return err == nil && found && len(addrs) > 0
}

func staleCondition(cond map[string]any, gen int64) bool {
	if gen == 0 {
		return false
	}
	observed, ok := nestedNumber(cond, "observedGeneration")
	if !ok || observed <= 0 {
		return false
	}
	return observed < gen
}

func nestedNumber(m map[string]any, key string) (int64, bool) {
	switch v := m[key].(type) {
	case int64:
		return v, true
	case float64:
		return int64(v), true
	}
	return 0, false
}

// why: staleGeneration means the conditions describe an older spec. Only claimed when
// the controller has written a generation, so kinds that never set it are exempt.
func staleGeneration(obj *unstructured.Unstructured) bool {
	gen := obj.GetGeneration()
	if gen == 0 || obj.GetDeletionTimestamp() != nil {
		return false
	}
	observed, found, err := unstructured.NestedInt64(obj.Object, "status", "observedGeneration")
	if err != nil || !found || observed <= 0 {
		return false
	}
	return observed < gen
}

var readyConditionTypes = []string{"Ready", "Available", "Established", "Synced"}

func isReadyConditionType(t string) bool {
	for _, want := range readyConditionTypes {
		if t == want {
			return true
		}
	}
	return false
}

func assessByConditions(obj *unstructured.Unstructured) (health, message, nativeStatus string) {
	if staleGeneration(obj) {
		return healthProgressing, "waiting for the controller to observe the current spec", ""
	}

	conds, ok, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil || !ok || len(conds) == 0 {
		return healthNotApplicable, "", ""
	}

	byType := map[string]map[string]any{}
	for _, c := range conds {
		cond, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if t, ok := cond["type"].(string); ok {
			byType[t] = cond
		}
	}

	gen := obj.GetGeneration()
	for _, want := range readyConditionTypes {
		cond, ok := byType[want]
		if !ok {
			continue
		}
		if staleCondition(cond, gen) {
			return healthProgressing, "waiting for the controller to observe the current spec", ""
		}
		status, _ := cond["status"].(string)
		msg, _ := cond["message"].(string)
		if msg == "" {
			msg, _ = cond["reason"].(string)
		}
		switch status {
		case "True":
			return healthHealthy, "", want + "=True"
		case "False":
			return healthDegraded, msg, want + "=False"
		default:
			return healthProgressing, msg, want + "=" + status
		}
	}

	return healthNotApplicable, "", ""
}

func mapHealth(code gitopshealth.HealthStatusCode) string {
	switch code {
	case gitopshealth.HealthStatusHealthy:
		return healthHealthy
	case gitopshealth.HealthStatusProgressing:
		return healthProgressing
	case gitopshealth.HealthStatusDegraded, gitopshealth.HealthStatusSuspended:
		return healthDegraded
	case gitopshealth.HealthStatusMissing:
		return healthUnhealthy
	default:
		return healthUnknown
	}
}

func resourceDetails(obj *unstructured.Unstructured, diagnosis map[string]any) string {
	details := map[string]any{}
	if len(diagnosis) > 0 {
		details["diagnosis"] = diagnosis
	}
	if status, ok := obj.Object["status"]; ok {
		details["status"] = status
	}
	if spec, ok := obj.Object["spec"].(map[string]any); ok {
		trimmed := make(map[string]any, len(spec))
		for k, v := range spec {
			if k == "template" {
				continue
			}
			trimmed[k] = v
		}
		details["spec"] = trimmed
	}
	if len(details) == 0 {
		return ""
	}

	if b, err := json.Marshal(details); err == nil && len(b) <= maxDetailsBytes {
		return string(b)
	}

	fallback := map[string]any{}
	if len(diagnosis) > 0 {
		fallback["diagnosis"] = diagnosis
	}
	if status, ok := obj.Object["status"]; ok {
		fallback["status"] = status
	}
	if len(fallback) > 0 {
		if b, err := json.Marshal(fallback); err == nil && len(b) <= maxDetailsBytes {
			return string(b)
		}
	}

	if len(diagnosis) > 0 {
		if b, err := json.Marshal(map[string]any{"diagnosis": diagnosis}); err == nil && len(b) <= maxDetailsBytes {
			return string(b)
		}
	}
	return ""
}
