package composite_errors

import (
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

// CompositeError is the runner-side representation of a structured error
// extracted from tool output (terraform, helm, actions). These are sent
// to the ctl-api as part of build/deploy reporting.
type CompositeError struct {
	OwnerType string         `json:"owner_type"`
	Severity     string         `json:"severity"`
	Summary      string         `json:"summary"`
	Detail       string         `json:"detail,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// FromGoError wraps a plain Go error as a single CompositeError with severity "critical".
func FromGoError(err error, ownerType string) []CompositeError {
	if err == nil {
		return nil
	}

	return []CompositeError{
		{
			OwnerType: ownerType,
			Severity:     "critical",
			Summary:      err.Error(),
		},
	}
}

// ToModels converts runner-side CompositeErrors to SDK model types for API reporting.
func ToModels(errs []CompositeError) []models.CompositeError {
	if len(errs) == 0 {
		return nil
	}

	out := make([]models.CompositeError, len(errs))
	for i, e := range errs {
		out[i] = models.CompositeError{
			OwnerType: e.OwnerType,
			Severity:     e.Severity,
			Summary:      e.Summary,
			Detail:       e.Detail,
			Metadata:     e.Metadata,
		}
	}
	return out
}
