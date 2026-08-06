package publish

import (
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	runnerairgap "github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/airgap/signals/publish/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type Signal struct {
	BundleID string                         `json:"bundle_id" validate:"required"`
	AppID    string                         `json:"app_id" validate:"required"`
	Runbooks []runnerairgap.RunbookTemplate `json:"runbooks,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(workflow.Context) error {
	if s.BundleID == "" {
		return fmt.Errorf("bundle_id is required")
	}
	if s.AppID == "" {
		return fmt.Errorf("app_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	if err := activities.AwaitUpdateBundleStatus(ctx, &activities.UpdateBundleStatusRequest{BundleID: s.BundleID, Status: app.AirgapBundleStatusPublishing, StatusDescription: "publishing bundle"}); err != nil {
		return fmt.Errorf("update bundle status: %w", err)
	}
	publishOpts := &workflow.ActivityOptions{RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 3}}
	if err := activities.AwaitPublishBundle(ctx, &activities.PublishBundleRequest{BundleID: s.BundleID, Runbooks: s.Runbooks}, publishOpts); err != nil {
		_ = activities.AwaitUpdateBundleStatus(ctx, &activities.UpdateBundleStatusRequest{BundleID: s.BundleID, Status: app.AirgapBundleStatusError, StatusDescription: err.Error()})
		return err
	}
	return activities.AwaitUpdateBundleStatus(ctx, &activities.UpdateBundleStatusRequest{BundleID: s.BundleID, Status: app.AirgapBundleStatusActive, StatusDescription: "bundle published and verified"})
}
