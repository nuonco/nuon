package flows

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/workflow"
)

func AppBranchUpdate(ctx workflow.Context, flw *app.Flow) ([]*app.FlowStep, error) {
	// flow step definition goes here
	return nil, nil
}
