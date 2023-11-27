package deprovision

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
)

const successNotificationTemplate string = `:white_check_mark: _successfully deprovisioned sandbox_ :white_check_mark:
• *s3-path*: %s
• *nuon-id*: _%s_
`

const errorNotificationTemplate string = `:rotating_light: _error deprovisioning sandbox_ :rotating_light:
• *s3-path*: %s
• *nuon-id*: _%s_
• *error-step*: _%s_

%s
`

type FinishRequest struct {
	DeprovisionRequest *installsv1.DeprovisionRequest

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
	dr := req.DeprovisionRequest
	s3Prefix := fmt.Sprintf("s3://%s/%s", req.InstallationsBucket, prefix.InstallPath(dr.OrgId, dr.AppId, dr.InstallId))
	notif := fmt.Sprintf(errorNotificationTemplate,
		s3Prefix,
		dr.InstallId,
		req.ErrorStep,
		req.ErrorMessage)

	return sender.Send(ctx, notif)
}

func (s *finisherImpl) sendSuccessNotification(ctx context.Context, req FinishRequest, sender notificationSender) error {
	dr := req.DeprovisionRequest
	s3Prefix := fmt.Sprintf("s3://%s/%s", req.InstallationsBucket, prefix.InstallPath(dr.OrgId, dr.AppId, dr.InstallId))
	notif := fmt.Sprintf(successNotificationTemplate,
		s3Prefix,
		dr.InstallId)

	return sender.Send(ctx, notif)
}
