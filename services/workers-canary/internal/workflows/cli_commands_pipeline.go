package workflows

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/workers-canary/internal/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *wkflow) execCLICommands(ctx workflow.Context, orgID string) error {
	var listComponentsResp activities.CLICommandResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CLICommand, &activities.CLICommandRequest{
		OrgID: orgID,
		Install: true,
		Json: true,
		Args: []string{
			"-j",
			"components",
			"list",
		},
	}, &listComponentsResp); err != nil {
		return fmt.Errorf("unable to execute list components: %w", err)
	}
	w.l.Info("list components", zap.Any("response", listComponentsResp))

	return nil
}
