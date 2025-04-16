package worker

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/poll"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

// @temporal-gen workflow
// @execution-timeout 24h
// @task-timeout 30s
func (w *Workflows) AwaitInstallStackVersionRun(ctx workflow.Context, sreq signals.RequestSignal) error {
	version, err := activities.AwaitGetInstallStackVersionByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install version")
	}

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	orgTyp, err := activities.AwaitGetOrgTypeByInstallID(ctx, sreq.ID)
	if err != nil {
		return err
	}

	if orgTyp == app.OrgTypeSandbox {
		l.Info("sandbox mode org")
		workflow.Sleep(ctx, time.Second*5)
		return nil
	}

	var run *app.InstallStackVersionRun
	if err := poll.Poll(ctx, w.v, poll.PollOpts{
		MaxTS:           workflow.Now(ctx).Add(time.Hour * 24),
		InitialInterval: time.Second * 15,
		MaxInterval:     time.Minute * 15,
		BackoffFactor:   2,
		Fn: func(ctx workflow.Context) error {
			run, err = activities.AwaitGetInstallStackVersionRunByVersionID(ctx, version.ID)
			return err
		},
	}); err != nil {
		statusactivities.AwaitPkgStatusUpdateInstallStackVersionStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: version.ID,
			Status: app.NewCompositeTemporalStatus(ctx, app.InstallStackVersionStatusActive, map[string]any{
				"err_message": err.Error(),
			}),
		})

		return errors.Wrap(err, "unable to get install stack run in time")
	}

	// successfully got a run
	l.Debug("successfully got run", zap.Any("data", run.Data))
	return nil
}
