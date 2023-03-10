package deprovision

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-workflows-meta/prefix"
	installsv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1"
)

const startNotificationTemplate string = `:construction: _started deprovisioning sandbox_ :construction:
• *s3-path*: %s
• *sandbox-name*: _%s_
• *sandbox-version*: _%s_
• *nuon-id*: _%s_
`

type StartRequest struct {
	DeprovisionRequest *installsv1.DeprovisionRequest

	InstallationsBucket string `json:"installations_bucket" validate:"required"`
}

func (s StartRequest) validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

type StartResponse struct{}

func (a *Activities) Start(ctx context.Context, req StartRequest) (StartResponse, error) {
	var resp StartResponse

	if err := a.sendStartNotification(ctx, req, a.sender); err != nil {
		return resp, fmt.Errorf("unable to send notification: %w", err)
	}

	return resp, nil
}

type starter interface {
	sendStartNotification(context.Context, StartRequest, notificationSender) error
	writeRequestFile(context.Context, StartRequest) error
}

var _ starter = (*starterImpl)(nil)

type starterImpl struct{}

func (s *starterImpl) sendStartNotification(ctx context.Context, req StartRequest, sender notificationSender) error {
	dr := req.DeprovisionRequest
	s3Prefix := fmt.Sprintf("s3://%s/%s", req.InstallationsBucket, prefix.InstallPath(dr.OrgId, dr.AppId, dr.InstallId))
	notif := fmt.Sprintf(startNotificationTemplate,
		s3Prefix,
		dr.SandboxSettings.Name,
		dr.SandboxSettings.Version,
		dr.InstallId)

	return sender.Send(ctx, notif)
}

func (s *starterImpl) writeRequestFile(ctx context.Context, req StartRequest) error {
	//TODO(jm): write out the request into the s3 prefix
	return nil
}
