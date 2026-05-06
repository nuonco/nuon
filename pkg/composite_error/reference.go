package composite_error

// ReferenceType enumerates the kinds of things a CompositeError can point at
// for dynamic resolution at read time.
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

// Reference points at another entity (or a URL) that the UI can dereference
// lazily. Keeps bulk material out of the persisted error row.
type Reference struct {
	Type  ReferenceType  `json:"type"`
	ID    string         `json:"id,omitempty"` // entity id, or url for *_url types
	Label string         `json:"label,omitempty"`
	Meta  map[string]any `json:"meta,omitempty"` // renderer-specific (e.g. {"start_line": 1421})
}

// Source captures the parser-input snippet that produced an error, plus
// identifiers for the parser that classified it. Kept small (capped) so
// debugging parser decisions doesn't require re-running the producing job.
type Source struct {
	ParserName    string `json:"parser_name,omitempty"`
	ParserVersion string `json:"parser_version,omitempty"`
	Snippet       string `json:"snippet,omitempty"` // capped at SourceSnippetMax bytes
	ExitCode      *int   `json:"exit_code,omitempty"`
	GoError       string `json:"go_error,omitempty"` // HumanError() output of the producing Go error
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
