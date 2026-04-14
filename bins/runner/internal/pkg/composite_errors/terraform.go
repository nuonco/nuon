package composite_errors

import (
	"bytes"
	"encoding/json"
)

// terraformJSONLine represents a single line of terraform -json output.
type terraformJSONLine struct {
	Level   string               `json:"@level"`
	Message string               `json:"@message"`
	Module  string               `json:"@module"`
	Type    string               `json:"type"`
	Diag    *terraformDiagnostic `json:"diagnostic,omitempty"`
}

type terraformDiagnostic struct {
	Severity string          `json:"severity"`
	Summary  string          `json:"summary"`
	Detail   string          `json:"detail"`
	Address  string          `json:"address"`
	Range    *terraformRange `json:"range,omitempty"`
}

type terraformRange struct {
	Filename string            `json:"filename"`
	Start    terraformPosition `json:"start"`
	End      terraformPosition `json:"end"`
}

type terraformPosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// ParseTerraformJSON parses terraform JSON output lines and extracts diagnostics
// into CompositeError values. Lines that are not diagnostics or error-level messages
// are skipped. Malformed JSON lines are silently ignored.
func ParseTerraformJSON(jsonOutput []byte, ownerType string) []CompositeError {
	if len(jsonOutput) == 0 {
		return nil
	}

	var errors []CompositeError

	lines := bytes.Split(jsonOutput, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var parsed terraformJSONLine
		if err := json.Unmarshal(line, &parsed); err != nil {
			continue
		}

		// Only process diagnostic lines or error-level messages.
		if parsed.Type != "diagnostic" && parsed.Level != "error" {
			continue
		}

		ce := CompositeError{
			OwnerType: ownerType,
			Metadata:     make(map[string]any),
		}

		if parsed.Diag != nil {
			ce.Summary = parsed.Diag.Summary
			ce.Detail = parsed.Diag.Detail
			ce.Severity = mapTerraformSeverity(parsed.Diag.Severity)

			if parsed.Diag.Address != "" {
				ce.Metadata["resource"] = parsed.Diag.Address
			}
			if parsed.Diag.Range != nil {
				ce.Metadata["file"] = parsed.Diag.Range.Filename
				ce.Metadata["line"] = parsed.Diag.Range.Start.Line
				ce.Metadata["column"] = parsed.Diag.Range.Start.Column
			}
		} else {
			// Error-level line without a diagnostic block.
			ce.Summary = parsed.Message
			ce.Severity = "critical"
		}

		// Drop empty metadata maps to keep output clean.
		if len(ce.Metadata) == 0 {
			ce.Metadata = nil
		}

		errors = append(errors, ce)
	}

	return errors
}

func mapTerraformSeverity(s string) string {
	switch s {
	case "error":
		return "critical"
	case "warning":
		return "warning"
	default:
		return s
	}
}
