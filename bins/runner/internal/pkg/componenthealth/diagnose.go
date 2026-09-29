package componenthealth

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const maxDiagnosisContainers = 6

func resourceDiagnosis(u *unstructured.Unstructured, health string, warn *warningEvent) map[string]any {
	if health == healthHealthy {
		return nil
	}

	diagnosis := map[string]any{}

	if warn != nil {
		diagnosis["event"] = map[string]any{
			"reason":  warn.reason,
			"message": warn.message,
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
