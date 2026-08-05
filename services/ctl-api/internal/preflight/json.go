package preflight

import (
	"encoding/json"
	"fmt"
	"io"
)

// The JSON shape is a deliberate DTO rather than tags on Result and Field.
// Field.Value holds raw config, secrets included, so marshalling it directly
// would make a leak one forgotten tag away. Here a secret has nowhere to go.
type jsonReport struct {
	Checks  []jsonCheck  `json:"checks"`
	Summary *jsonSummary `json:"summary,omitempty"`
}

type jsonCheck struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// Absent for --list, which resolves fields without running anything.
	Status Status      `json:"status,omitempty"`
	Detail string      `json:"detail,omitempty"`
	Fields []jsonField `json:"fields"`
}

type jsonField struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Secret   bool   `json:"secret"`
	// Set separates "configured" from "empty" without publishing the value,
	// which is the only thing callers need for a secret.
	Set   bool   `json:"set"`
	Value string `json:"value,omitempty"`
}

type jsonSummary struct {
	Passed  int `json:"passed"`
	Warned  int `json:"warned"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// WriteJSONResults emits a run as JSON and returns the process exit code, using
// the same rule as the table: only a failure is non-zero.
func WriteJSONResults(w io.Writer, results []Result) int {
	summary := summarize(results)
	if err := writeJSON(w, jsonReport{Checks: toJSONChecks(results), Summary: &summary}); err != nil {
		fmt.Fprintf(w, "unable to encode results: %v\n", err)

		return 2
	}

	if summary.Failed > 0 {
		return 1
	}

	return 0
}

// WriteJSONChecks emits the check catalogue as JSON. No summary: nothing ran.
func WriteJSONChecks(w io.Writer, results []Result) error {
	return writeJSON(w, jsonReport{Checks: toJSONChecks(results)})
}

func writeJSON(w io.Writer, report jsonReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(report)
}

func toJSONChecks(results []Result) []jsonCheck {
	checks := make([]jsonCheck, 0, len(results))
	for _, r := range results {
		checks = append(checks, jsonCheck{
			Name:        r.Name,
			Description: r.Description,
			Status:      r.Status,
			Detail:      r.Detail,
			Fields:      toJSONFields(r.Fields),
		})
	}

	return checks
}

func toJSONFields(fields []Field) []jsonField {
	out := make([]jsonField, 0, len(fields))
	for _, f := range fields {
		jf := jsonField{
			Name:     f.Name,
			Required: f.Required,
			Secret:   f.Secret,
			Set:      f.Value != "",
		}
		if !f.Secret {
			jf.Value = f.Value
		}
		out = append(out, jf)
	}

	return out
}

func summarize(results []Result) jsonSummary {
	var s jsonSummary
	for _, r := range results {
		switch r.Status {
		case StatusPass:
			s.Passed++
		case StatusWarn:
			s.Warned++
		case StatusFail:
			s.Failed++
		case StatusSkipped:
			s.Skipped++
		}
	}

	return s
}
