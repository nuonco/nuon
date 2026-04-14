package models

type CompositeError struct {
	OwnerType string         `json:"owner_type"`
	Severity  string         `json:"severity"`
	Summary   string         `json:"summary"`
	Detail    string         `json:"detail,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type ReportCompositeErrorsRequest struct {
	Errors []CompositeError `json:"errors"`
}
