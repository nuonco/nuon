package polldependencies

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "poll_dependencies"

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

type Signal struct {
	AppID string `json:"app_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.AppID == "" {
		return errors.New("app_id is required")
	}

	_, err := activities.AwaitGetByAppID(ctx, s.AppID)
	if err != nil {
		return errors.Wrap(err, "app not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	for {
		currentApp, err := activities.AwaitGetByAppID(ctx, s.AppID)
		if err != nil {
			if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
				AppID:             s.AppID,
				Status:            app.AppStatusError,
				StatusDescription: "unable to get app from database",
			}); updateErr != nil {
				workflow.GetLogger(ctx).Error("failed to update app status", updateErr)
			}
			statusactivities.AwaitUpdateAppStatusV2(ctx, statusactivities.UpdateAppStatusV2Request{
				AppID:             s.AppID,
				Status:            app.AppStatusError,
				StatusDescription: "unable to get app from database",
			})
			return errors.Wrap(err, "unable to get app from database")
		}

		if currentApp.Org.Status == "active" {
			return nil
		}

		if currentApp.Org.Status == "error" {
			// TODO(sdboyer) remove transitive error status propagation
			if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
				AppID:             s.AppID,
				Status:            "error",
				StatusDescription: "org is in error state",
			}); err != nil {
				return errors.Wrap(err, "unable to update status")
			}
			statusactivities.AwaitUpdateAppStatusV2(ctx, statusactivities.UpdateAppStatusV2Request{
				AppID:             s.AppID,
				Status:            app.AppStatusError,
				StatusDescription: "org is in error state",
			})
		}

		workflow.Sleep(ctx, defaultPollTimeout)
	}
}
