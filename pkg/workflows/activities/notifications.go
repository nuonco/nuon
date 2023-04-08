package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/sender"
	sharedactivitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1/activities/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultSendNotificationTimeout time.Duration = time.Second * 5
)

// SendNotification is a method that can be called from a workflow to send a notification using an activity
func SendNotification(ctx workflow.Context, req *sharedactivitiesv1.SendNotificationRequest) error {
	l := workflow.GetLogger(ctx)
	l.Debug("executing send notification activity")

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultSendNotificationTimeout,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	var resp sharedactivitiesv1.SendNotificationResponse
	fut := workflow.ExecuteActivity(ctx, "SendNotification", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return fmt.Errorf("unable to send notification: %w", err)
	}

	return nil
}

func (a *Activities) SendNotification(ctx context.Context, req *sharedactivitiesv1.SendNotificationRequest) (*sharedactivitiesv1.SendNotificationResponse, error) {
	var resp sharedactivitiesv1.SendNotificationResponse

	l := activity.GetLogger(ctx)
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("unable to validate message: %w", err)
	}

	if req.DryRun {
		l.Info("sending notification: "+req.Notification, "dry-run", true)
		return &resp, nil
	}

	sndr, err := sender.NewSlackSender(req.SlackWebhookUrl)
	if err != nil {
		return nil, fmt.Errorf("unable create slack sender: %w", err)
	}

	if err := sndr.Send(ctx, req.Notification); err != nil {
		return nil, fmt.Errorf("unable to send notification: %w", err)
	}

	return &resp, nil
}
