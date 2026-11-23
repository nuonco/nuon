package state

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	state "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func Regenerate(ctx workflow.Context, req *state.ExecuteRegenerationRequest) (*state.ExecuteRegenerationResponse, error) {
	activityReq := *req
	activityReq.CachedState, activityReq.MetricsWriter = nil, nil
	resp, err := installactivities.AwaitRegenerateInstallState(ctx, &activityReq)
	if err != nil {
		return nil, errors.Wrap(err, "unable to regenerate install state")
	}
	if len(resp.UpdatedPartials) > 0 {
		if err := installactivities.AwaitRenderInstallLabels(ctx, &installactivities.RenderInstallLabelsRequest{InstallID: req.InstallID}); err != nil {
			workflow.GetLogger(ctx).Warn("unable to render install label templates", "install_id", req.InstallID, "error", err)
		}
	}
	return resp, nil
}
