package installs

import (
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func workflowIDFromResp(resp *models.AppWorkflowResponse) string {
	if resp == nil {
		return ""
	}
	return resp.WorkflowID
}

type actionResult struct {
	InstallID  string `json:"install_id,omitempty"`
	ID         string `json:"id,omitempty"`
	WorkflowID string `json:"workflow_id,omitempty"`
	Status     string `json:"status"`
}

func printActionResult(asJSON bool, humanMsg string, r actionResult) {
	if asJSON {
		ui.PrintJSON(r)
		return
	}
	ui.PrintLn(humanMsg)
}
