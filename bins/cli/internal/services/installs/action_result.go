package installs

import (
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// workflowIDFromResp safely extracts a workflow id from an app workflow response
// that may be nil.
func workflowIDFromResp(resp *models.AppWorkflowResponse) string {
	if resp == nil {
		return ""
	}
	return resp.WorkflowID
}

// actionResult is the JSON payload emitted by mutating install commands so that
// --output json / --output agent callers receive a machine-readable
// confirmation instead of an empty envelope.
type actionResult struct {
	InstallID  string `json:"install_id,omitempty"`
	ID         string `json:"id,omitempty"`
	WorkflowID string `json:"workflow_id,omitempty"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
}

// printActionResult renders a mutating command's outcome: a JSON payload when
// asJSON is set (raw JSON or agent envelope, handled by ui.PrintJSON), otherwise
// a human status line. The same human message is carried in the JSON payload's
// "message" field so machine callers get both the stable status token and the
// human-readable prose.
func printActionResult(asJSON bool, humanMsg string, r actionResult) {
	if asJSON {
		r.Message = humanMsg
		ui.PrintJSON(r)
		return
	}
	ui.PrintLn(humanMsg)
}
