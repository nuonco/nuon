package componenthealth

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// maxDiagnosisContainers bounds how many container statuses travel with a
// single resource — enough to explain a failure without shipping a large pod.
const maxDiagnosisContainers = 6

// resourceDiagnosis captures the "why" behind a resource that is not healthy:
// the controller-side warning event, and for pods the restart counts and
// termination reasons that name the actual failure (OOMKilled,
// ImagePullBackOff, CrashLoopBackOff). Returns nil for healthy resources.
//
// Everything here comes from objects the engine already lists, so diagnosis
// costs no extra API calls and needs no permissions beyond the read access the
// engine already has. A bounded log tail would need pods/log and is
// deliberately left out.
//
// The runner cannot know whether a verdict transitioned — it is stateless and
// the verdict is computed server-side — so diagnosis travels on every report
// for any non-healthy resource, and the evaluator keeps the copy belonging to
// the resource that caused a transition.
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
