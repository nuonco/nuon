package componenthealth

import (
	"encoding/json"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/componenthealth/gitopshealth"
)

const (
	healthHealthy     = "healthy"
	healthProgressing = "progressing"
	healthDegraded    = "degraded"
	healthUnhealthy   = "unhealthy"
	healthUnknown     = "unknown"
	// healthNotApplicable is for a resource we successfully read but which
	// exposes no health signal at all. Distinct from unknown, which means we
	// could not tell — claiming unknown here would report a permanent absence of
	// information for something that simply has none to give.
	healthNotApplicable = "not-applicable"

	// maxDetailsBytes bounds the per-resource status blob the runner sends.
	maxDetailsBytes = 8 * 1024
)

// assessResource derives the generic health verdict for a live object using
// Argo's gitops-engine assessment, preserving the native status string.
func assessResource(obj *unstructured.Unstructured) (health, message, nativeStatus string) {
	hs, err := gitopshealth.GetResourceHealth(obj, nil)
	if err != nil {
		return healthUnknown, "", ""
	}
	if hs == nil {
		// The library knows ~10 kinds and returns nil for the rest. Rather than
		// call every CRD unknown forever, try the convention most controllers
		// follow, then admit the resource has no signal.
		return assessByConditions(obj)
	}

	health = mapHealth(hs.Status)
	if (health == healthHealthy || health == healthProgressing) && !staleGeneration(obj) {
		if reason, msg, ok := conditionFailure(obj); ok {
			return healthDegraded, msg, reason
		}
	}

	msg := hs.Message
	if msg == "" {
		msg = explainVerdict(obj, hs.Status)
	}
	return health, msg, string(hs.Status)
}

// conditionFailure finds a current failure the object names in its own
// conditions. Every per-kind check upstream reads a narrow slice of status and
// treats it as the whole truth: the Deployment check consults only
// Progressing/ProgressDeadlineExceeded and never ReplicaFailure, and the HPA
// check returns on the first condition that matches anything, so
// AbleToScale=True masked ScalingActive=False/FailedGetResourceMetric. Both
// reported healthy while the object said otherwise, which is what pushed the
// only usable signal into Warning events.
//
// Reading every condition instead keeps the verdict a pure function of one read
// of status, so it needs no expiry and clears the moment the controller does.
func conditionFailure(obj *unstructured.Unstructured) (reason, message string, found bool) {
	conds, ok, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil || !ok {
		return "", "", false
	}

	for _, c := range conds {
		cond, ok := c.(map[string]any)
		if !ok {
			continue
		}
		condReason, _ := cond["reason"].(string)
		if !failureReason(condReason) {
			continue
		}
		// A ready-style condition that currently reads True is asserting health;
		// a failure reason left behind on it describes something already over.
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

// failureReason reports whether a condition reason names a failure rather than a
// state. Polarity cannot be read from status: ReplicaFailure=True and
// ScalingActive=False both mean broken, so the reason is the only consistent
// signal. Follows the same naming idiom the vendored pod check already trusts
// for container waiting reasons.
func failureReason(reason string) bool {
	switch {
	case reason == "":
		return false
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

// explainVerdict fills in a message the library leaves blank. A verdict with no
// message is the least actionable thing health can show — an Ingress waiting on
// a load balancer address reported "progressing" and nothing else for 15 hours,
// while the reason sat in the status it had already read.
func explainVerdict(obj *unstructured.Unstructured, status gitopshealth.HealthStatusCode) string {
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

// staleGeneration reports a controller that has not yet processed the object's
// current spec, so its conditions describe the previous one. kstatus treats
// this as in-progress by definition; without it a fresh spec is graded against
// stale conditions. Only claimed when the controller has demonstrably written a
// generation and it is behind, so kinds that never set the field are unaffected.
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

// readyConditionTypes are the condition types controllers conventionally use to
// mean "this object is serving". Ordered by preference.
var readyConditionTypes = []string{"Ready", "Available", "Established", "Synced"}

func isReadyConditionType(t string) bool {
	for _, want := range readyConditionTypes {
		if t == want {
			return true
		}
	}
	return false
}

// assessByConditions reads the status.conditions convention. A True ready-style
// condition is healthy, False is degraded with its message, Unknown is
// progressing. An object with no such condition is not-applicable: we read it
// successfully and it has nothing to say about its own health.
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

	for _, want := range readyConditionTypes {
		cond, ok := byType[want]
		if !ok {
			continue
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

// resourceDetails returns a bounded JSON summary for the detail view. Spec is
// included because status alone is uninformative for e.g. a ClusterIP Service,
// and diagnosis outranks it since the evaluator copies it onto transitions.
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

	// spec pushed it over the cap — fall back to status (plus diagnosis) only.
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

	// Still too large: keep the diagnosis, drop the raw status.
	if len(diagnosis) > 0 {
		if b, err := json.Marshal(map[string]any{"diagnosis": diagnosis}); err == nil && len(b) <= maxDetailsBytes {
			return string(b)
		}
	}
	return ""
}
