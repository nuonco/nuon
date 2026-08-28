package publish

import (
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/signals/publish/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type Signal struct {
	PackageID string `json:"package_id" validate:"required"`
	AppID     string `json:"app_id" validate:"required"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(workflow.Context) error {
	if s.PackageID == "" {
		return fmt.Errorf("package_id is required")
	}
	if s.AppID == "" {
		return fmt.Errorf("app_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	if err := activities.AwaitUpdateBundleStatus(ctx, &activities.UpdateBundleStatusRequest{PackageID: s.PackageID, Status: app.ReleasePackageStatusPublishing, StatusDescription: "publishing package"}); err != nil {
		return fmt.Errorf("update package status: %w", err)
	}
	publishOpts := &workflow.ActivityOptions{RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 3}}
	if err := activities.AwaitPublishBundle(ctx, &activities.PublishBundleRequest{PackageID: s.PackageID}, publishOpts); err != nil {
		_ = activities.AwaitUpdateBundleStatus(ctx, &activities.UpdateBundleStatusRequest{PackageID: s.PackageID, Status: app.ReleasePackageStatusError, StatusDescription: err.Error()})
		return err
	}
	return activities.AwaitUpdateBundleStatus(ctx, &activities.UpdateBundleStatusRequest{PackageID: s.PackageID, Status: app.ReleasePackageStatusActive, StatusDescription: "package published and verified"})
}
