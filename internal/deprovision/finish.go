package deprovision

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
)

const successNotificationTemplate string = `:white_check_mark: _successfully deprovisioned sandbox_ :white_check_mark:
• *s3-path*: %s
• *sandbox-name*: _%s_
• *sandbox-version*: _%s_
• *nuon-id*: _%s_
`

const errorNotificationTemplate string = `:rotating_light: _error deprovisioning sandbox_ :rotating_light:
• *s3-path*: %s
• *sandbox-name*: _%s_
• *sandbox-version*: _%s_
• *nuon-id*: _%s_
• *error-step*: _%s_

%s
`

type FinishRequest struct {
	DeprovisionRequest

	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message"`
	ErrorStep    string `json:"error_step"`

	InstallationsBucket string `json:"installations_bucket" validate:"required"`
}

func (f FinishRequest) validate() error {
	validate := validator.New()
	return validate.Struct(f)
}

type FinishResponse struct{}

func (a *Activities) FinishDeprovision(ctx context.Context, req FinishRequest) (FinishResponse, error) {
	var resp FinishResponse

	fn := a.sendSuccessNotification
	if !req.Success {
		fn = a.sendErrorNotification
	}

	if err := fn(ctx, req, a.sender); err != nil {
		return resp, fmt.Errorf("unable to send finish notification: %w", err)
	}

	return resp, nil
}

type finisher interface {
	sendSuccessNotification(context.Context, FinishRequest, notificationSender) error
	sendErrorNotification(context.Context, FinishRequest, notificationSender) error
}

var _ finisher = (*finisherImpl)(nil)

type finisherImpl struct{}

func (s *finisherImpl) sendErrorNotification(ctx context.Context, req FinishRequest, sender notificationSender) error {
	s3Prefix := getS3Prefix(req.InstallationsBucket, req.OrgID, req.AppID, req.InstallID)
	notif := fmt.Sprintf(errorNotificationTemplate,
		s3Prefix,
		req.SandboxSettings.Name,
		req.SandboxSettings.Version,
		req.InstallID,
		req.ErrorStep,
		req.ErrorMessage)

	return sender.Send(ctx, notif)
}

func (s *finisherImpl) sendSuccessNotification(ctx context.Context, req FinishRequest, sender notificationSender) error {
	s3Prefix := getS3Prefix(req.InstallationsBucket, req.OrgID, req.AppID, req.InstallID)
	notif := fmt.Sprintf(successNotificationTemplate,
		s3Prefix,
		req.SandboxSettings.Name,
		req.SandboxSettings.Version,
		req.InstallID)

	return sender.Send(ctx, notif)
}
