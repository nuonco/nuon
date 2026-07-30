package componenthealth

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/componenthealth/gitopshealth"
)

const (
	healthHealthy     = "healthy"
	healthProgressing = "progressing"
	healthDegraded    = "degraded"
	healthUnhealthy   = "unhealthy"
	healthUnknown     = "unknown"

	// maxDetailsBytes bounds the per-resource status blob the runner sends.
	maxDetailsBytes = 8 * 1024
)

// assessResource derives the generic health verdict for a live object using
// Argo's gitops-engine assessment, preserving the native status string.
func assessResource(obj *unstructured.Unstructured) (health, message, nativeStatus string) {
	hs, err := gitopshealth.GetResourceHealth(obj, nil)
	if err != nil || hs == nil {
		return healthUnknown, "", ""
	}
	return mapHealth(hs.Status), hs.Message, string(hs.Status)
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
