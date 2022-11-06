package provision

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-sender"
	"github.com/powertoolsdev/go-uploader"
)

type FinishRequest struct {
	ProvisionRequest

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

func (a *ProvisionActivities) Finish(ctx context.Context, req FinishRequest) (FinishResponse, error) {
	var resp FinishResponse

	if err := req.validate(); err != nil {
		return resp, err
	}

	fn := a.sendSuccessNotification
	if !req.Success {
		fn = a.sendErrorNotification
	}

	if err := fn(ctx, req, a.sender); err != nil {
		return resp, fmt.Errorf("unable to send finish notification: %w", err)
	}

	// write status file to S3
	s3Prefix := getInstallationPrefix(
		req.ProvisionRequest.OrgID,
		req.ProvisionRequest.AppID,
		req.ProvisionRequest.InstallID)
	statusFileContents := StatusFileContents{
		Status:       "Finished",
		ErrorStep:    req.ErrorStep,
		ErrorMessage: req.ErrorMessage,
	}
	uploadClient := uploader.NewS3Uploader(req.InstallationsBucket, s3Prefix)
	if err := a.writeStatusFile(ctx, uploadClient, statusFileContents); err != nil {
		return resp, fmt.Errorf("unable to upload status file to s3: %w", err)
	}

	return resp, nil
}

type finisher interface {
	sendSuccessNotification(context.Context, FinishRequest, sender.NotificationSender) error
	sendErrorNotification(context.Context, FinishRequest, sender.NotificationSender) error
}

var _ finisher = (*finisherImpl)(nil)

type finisherImpl struct{}

const errorNotificationTemplate string = `:rotating_light: _error provisioning sandbox_ :rotating_light:
• *s3-path*: %s
• *sandbox-name*: _%s_
• *sandbox-version*: _%s_
• *nuon-id*: _%s_
• *error-step*: _%s_

%s
`

func (s *finisherImpl) sendErrorNotification(ctx context.Context, req FinishRequest, sender sender.NotificationSender) error {
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

const successNotificationTemplate string = `:white_check_mark: _successfully provisioned sandbox_ :white_check_mark:
• *s3-path*: %s
• *sandbox-name*: _%s_
• *sandbox-version*: _%s_
• *nuon-id*: _%s_
• *runner-id*: _%s_
• *org-id*: _%s_
`

func (s *finisherImpl) sendSuccessNotification(ctx context.Context, req FinishRequest, sender sender.NotificationSender) error {
	s3Prefix := getS3Prefix(req.InstallationsBucket, req.OrgID, req.AppID, req.InstallID)
	notif := fmt.Sprintf(successNotificationTemplate,
		s3Prefix,
		req.SandboxSettings.Name,
		req.SandboxSettings.Version,
		req.InstallID,
		req.InstallID,
		req.OrgID)

	return sender.Send(ctx, notif)
}
