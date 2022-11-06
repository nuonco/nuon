package provision

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-sender"
	"github.com/powertoolsdev/go-uploader"
)

const (
	requestFilename           = "request.json"
	statusFilename            = "status.json"
	startNotificationTemplate = `:package: _started provisioning sandbox_ :package:
• *s3-path*: s3://%s/%s
• *sandbox-name*: _%s_
• *sandbox-version*: _%s_
• *role*: _%s_
• *nuon-id*: _%s_
`
)

type StatusFileContents struct {
	Status       string `json:"status" validate:"required"`
	ErrorStep    string `json:"error_step,omitempty" validate:"required"`
	ErrorMessage string `json:"error_message,omitempty" validate:"required"`
}

type StartWorkflowRequest struct {
	OrgID               string `json:"org_id" validate:"required"`
	AppID               string `json:"app_id" validate:"required"`
	InstallID           string `json:"install_id" validate:"required"`
	InstallationsBucket string `json:"installations_bucket" validate:"required"`

	ProvisionRequest ProvisionRequest `json:"provision_request" validate:"required"`
}

func (p StartWorkflowRequest) validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type StartWorkflowResponse struct{}

var errNoValidSender error = fmt.Errorf("no sender specified")

// sendStartNotification sends the start notification via the configured sender
func (n *starterImpl) sendStartNotification(ctx context.Context, req StartWorkflowRequest) error {
	if n.sender == nil {
		return errNoValidSender
	}

	prefix := getInstallationPrefix(req.OrgID, req.AppID, req.InstallID)

	msg := fmt.Sprintf(startNotificationTemplate, req.InstallationsBucket, prefix,
		req.ProvisionRequest.SandboxSettings.Name,
		req.ProvisionRequest.SandboxSettings.Version,
		req.ProvisionRequest.AccountSettings.AwsRoleArn,
		req.InstallID,
	)

	return n.sender.Send(ctx, msg)
}

func (n *starterImpl) writeRequestFile(ctx context.Context, client s3BlobUploader, req ProvisionRequest) error {
	byts, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// upload file
	if err := client.UploadBlob(ctx, byts, requestFilename); err != nil {
		return fmt.Errorf("unable to upload request file to s3: %s", err)
	}

	return nil
}

func (n *starterImpl) writeStatusFile(ctx context.Context, client s3BlobUploader, statusFileContents StatusFileContents) error {
	byts, err := json.Marshal(statusFileContents)
	if err != nil {
		return err
	}

	// upload file
	if err := client.UploadBlob(ctx, byts, statusFilename); err != nil {
		return fmt.Errorf("unable to upload status file to s3: %s", err)
	}

	return nil
}

func (a *ProvisionActivities) StartWorkflow(ctx context.Context, req StartWorkflowRequest) (StartWorkflowResponse, error) {
	var resp StartWorkflowResponse

	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	// send start notification
	if err := a.sendStartNotification(ctx, req); err != nil {
		return resp, fmt.Errorf("unable to send notification: %w", err)
	}

	// write request file to S3
	s3Prefix := getInstallationPrefix(
		req.ProvisionRequest.OrgID,
		req.ProvisionRequest.AppID,
		req.ProvisionRequest.InstallID)
	uploadClient := uploader.NewS3Uploader(req.InstallationsBucket, s3Prefix)
	if err := a.writeRequestFile(ctx, uploadClient, req.ProvisionRequest); err != nil {
		return resp, fmt.Errorf("unable to upload request file to s3: %w", err)
	}
	statusFileContents := StatusFileContents{
		Status: "Started",
	}
	if err := a.writeStatusFile(ctx, uploadClient, statusFileContents); err != nil {
		return resp, fmt.Errorf("unable to upload status file to s3: %w", err)
	}

	return resp, nil
}

type starter interface {
	sendStartNotification(context.Context, StartWorkflowRequest) error
	writeRequestFile(context.Context, s3BlobUploader, ProvisionRequest) error
	writeStatusFile(context.Context, s3BlobUploader, StatusFileContents) error
}

type starterImpl struct {
	sender sender.NotificationSender
}

type s3BlobUploader interface {
	UploadBlob(context.Context, []byte, string) error
}
