package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// updateBuildStatus updates the build status. An optional CompositeErrorData
// can be passed as the final argument to also persist a typed structured
// error against the build's composite_error column.
func (w *Workflows) updateBuildStatus(ctx workflow.Context, bldID string, status app.ComponentBuildStatus, statusDescription string, ces ...*compositeerrors.CompositeErrorData) {
	l := workflow.GetLogger(ctx)
	var ce *compositeerrors.CompositeErrorData
	if len(ces) > 0 {
		ce = ces[0]
	}
	err := activities.AwaitUpdateBuildStatus(ctx, activities.UpdateBuildStatus{
		BuildID:           bldID,
		Status:            status,
		StatusDescription: statusDescription,
		CompositeError:    ce,
	})
	if err != nil {
		l.Error("unable to update build status",
			zap.String("build-id", bldID),
			zap.Error(err))
		return
	}

	err = statusactivities.AwaitUpdateBuildStatusV2(ctx, statusactivities.UpdateBuildStatusV2{
		BuildID:           bldID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err != nil {
		l.Error("unable to update build status v2",
			zap.String("build-id", bldID),
			zap.Error(err))
		return
	}

}
