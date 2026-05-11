package composite_error

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// ReferenceType enumerates the kinds of things a CompositeError can point at
// for dynamic resolution at read time. Lets the persisted error stay small
// while letting the UI dereference bulk material lazily.
type ReferenceType string

const (
	RefTypeLogStream                ReferenceType = "log_stream"
	RefTypeRunnerJobExecution       ReferenceType = "runner_job_execution"
	RefTypeRunnerJobExecutionResult ReferenceType = "runner_job_execution_result"
	RefTypeTerraformPlanResult      ReferenceType = "terraform_plan_result"
	RefTypeWorkflowStep             ReferenceType = "workflow_step"
	RefTypeInstallDeploy            ReferenceType = "install_deploy"
	RefTypeComponentBuild           ReferenceType = "component_build"
	RefTypeDocURL                   ReferenceType = "doc_url"
	RefTypeRunbookURL               ReferenceType = "runbook_url"
)

// Reference points at another entity (or a URL) the UI can dereference at
// read time. Meta is renderer-specific (e.g. {"start_line": 1421}).
type Reference struct {
	Type  ReferenceType  `json:"type"`
	ID    string         `json:"id,omitempty"`
	Label string         `json:"label,omitempty"`
	Meta  map[string]any `json:"meta,omitempty"`
}

// References is a slice of Reference that satisfies sql.Scanner and
// driver.Valuer so it can be stored directly as a JSONB column on the
// composite_errors row.
type References []Reference

func (r References) Value() (driver.Value, error) {
	if len(r) == 0 {
		return nil, nil
	}
	return json.Marshal(r)
}

func (r *References) Scan(v any) error {
	if v == nil {
		return nil
	}
	bytes, ok := v.([]byte)
	if !ok {
		return errors.New("composite_error.References: invalid scan type")
	}
	if len(bytes) == 0 || string(bytes) == "null" {
		return nil
	}
	return json.Unmarshal(bytes, r)
}

func (References) GormDataType() string { return "jsonb" }

// Source captures the parser-input snippet that produced an error, plus
// identifiers for the parser that classified it. Kept small (capped) so that
// debugging parser decisions doesn't require re-running the producing job.
type Source struct {
	ParserName    string `json:"parser_name,omitempty"`
	ParserVersion string `json:"parser_version,omitempty"`
	Snippet       string `json:"snippet,omitempty"`
	ExitCode      *int   `json:"exit_code,omitempty"`
	GoError       string `json:"go_error,omitempty"`
}

// SourceSnippetMax caps how much raw input we persist on a CompositeError row.
const SourceSnippetMax = 8 * 1024

// CapSnippet truncates s to at most SourceSnippetMax bytes, appending an
// ellipsis marker when truncation occurs.
func CapSnippet(s string) string {
	if len(s) <= SourceSnippetMax {
		return s
	}
	const marker = "\n…[truncated]"
	return s[:SourceSnippetMax-len(marker)] + marker
}

func (s Source) Value() (driver.Value, error) {
	if s == (Source{}) {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *Source) Scan(v any) error {
	if v == nil {
		return nil
	}
	bytes, ok := v.([]byte)
	if !ok {
		return errors.New("composite_error.Source: invalid scan type")
	}
	if len(bytes) == 0 || string(bytes) == "null" {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

func (Source) GormDataType() string { return "jsonb" }
