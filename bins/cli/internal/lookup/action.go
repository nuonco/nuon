package lookup

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

// ActionWorkflowID resolves an action ID or name within an app.
func ActionWorkflowID(ctx context.Context, apiClient nuon.Client, appID, actionIDOrName string) (string, error) {
	if actionIDOrName == "" {
		return "", &ui.CLIUserError{
			Msg: "an action ID or name is required, pass the --action-workflow-id flag",
		}
	}

	action, err := apiClient.GetAppActionWorkflow(ctx, appID, actionIDOrName)
	if nuon.IsNotFound(err) {
		return "", &ui.CLIUserError{
			Msg: fmt.Sprintf("action %q not found in app %s", actionIDOrName, appID),
		}
	}
	if err != nil {
		return "", err
	}

	return action.ID, nil
}
