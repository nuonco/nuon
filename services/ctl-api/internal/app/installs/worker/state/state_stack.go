package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/stacks"
	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getStackStatePartial(ctx workflow.Context, installID string) (*state.InstallStackState, error) {
	stack, err := activities.AwaitGetInstallStackStateByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get stack")
	}

	return w.toInstallStackState(stack), nil
}

func (h *Workflows) toInstallStackState(stack *app.InstallStack) *state.InstallStackState {
	if stack == nil || len(stack.InstallStackVersions) < 1 {
		return nil
	}

	is := state.NewInstallStackState()
	is.Populated = true

	version := stack.InstallStackVersions[0]
	is.QuickLinkURL = version.QuickLinkURL
	is.TemplateURL = version.TemplateURL
	is.TemplateJSON = string(version.Contents)
	is.Checksum = version.Checksum
	is.Status = string(version.Status.Status)

	outputData, err := stacks.DecodeAWSStackOutputData(stack.InstallStackOutputs.Data)
	if err != nil {
		return nil
	}
	is.Outputs = outputData

	return is
}
