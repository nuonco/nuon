package componenthealth

import (
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// maxDiagnosisContainers bounds how many container statuses travel with a
// single resource — enough to explain a failure without shipping a large pod.
const maxDiagnosisContainers = 6

// resourceDiagnosis explains an unhealthy resource from objects already listed.
// Rides every non-healthy report: the stateless runner can't know what transitioned.
func resourceDiagnosis(u *unstructured.Unstructured, health string, warn *warningEvent) map[string]any {
	if health == healthHealthy {
		return nil
	}

	diagnosis := map[string]any{}

	if warn != nil {
		event := map[string]any{
			"reason":  warn.reason,
			"message": warn.message,
		}
		if !warn.at.IsZero() {
			event["observed_at"] = warn.at.UTC().Format(time.RFC3339)
		}
		diagnosis["event"] = event
		if warn.source.valid() {
			diagnosis["source"] = warn.source.details()
		}
		if len(warn.ownerPath) > 1 {
			path := make([]map[string]any, 0, len(warn.ownerPath))
			for _, ref := range warn.ownerPath {
				path = append(path, ref.details())
			}
			diagnosis["owner_path"] = path
		}
	}

	if containers := containerDiagnosis(u); len(containers) > 0 {
		diagnosis["containers"] = containers
	}

	if len(diagnosis) == 0 {
		return nil
	}
	return diagnosis
}

// containerDiagnosis summarises the failing containers of a pod. Returns nil
// for any other kind, and skips containers that are running cleanly.
func containerDiagnosis(u *unstructured.Unstructured) []map[string]any {
	if u.GetKind() != "Pod" {
		return nil
	}

	statuses, ok, _ := unstructured.NestedSlice(u.Object, "status", "containerStatuses")
	if !ok {
		return nil
	}

	out := make([]map[string]any, 0, len(statuses))
	for _, raw := range statuses {
		cs, ok := raw.(map[string]any)
		if !ok {
			continue
		}

		entry := map[string]any{}
		if name, ok, _ := unstructured.NestedString(cs, "name"); ok && name != "" {
			entry["name"] = name
		}
		if restarts, ok, _ := unstructured.NestedInt64(cs, "restartCount"); ok && restarts > 0 {
			entry["restart_count"] = restarts
		}
		if reason, ok, _ := unstructured.NestedString(cs, "state", "waiting", "reason"); ok && reason != "" {
			entry["waiting_reason"] = reason
			if msg, ok, _ := unstructured.NestedString(cs, "state", "waiting", "message"); ok && msg != "" {
				entry["waiting_message"] = msg
			}
		}
		if reason, ok, _ := unstructured.NestedString(cs, "lastState", "terminated", "reason"); ok && reason != "" {
			entry["last_termination_reason"] = reason
			if code, ok, _ := unstructured.NestedInt64(cs, "lastState", "terminated", "exitCode"); ok {
				entry["last_termination_exit_code"] = code
			}
		}

		// Only the name means the container has nothing wrong to report.
		if len(entry) <= 1 {
			continue
		}
		out = append(out, entry)
		if len(out) == maxDiagnosisContainers {
			break
		}
	}

	if len(out) == 0 {
		return nil
	}
	return out
}
