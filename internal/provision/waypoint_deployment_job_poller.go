package provision

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-uploader"
	"github.com/powertoolsdev/go-waypoint/job"
	"go.temporal.io/sdk/activity"
)

const eventFilename = "events.json"

// for mocking purposes
var (
	osReadFile   = os.ReadFile
	osRemoveFile = os.Remove
)

type PollWaypointDeploymentJobRequest struct {
	OrgID                string `json:"org_id" validate:"required"`
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`

	BucketName   string `json:"bucket_name" validate:"required"`
	BucketPrefix string `json:"bucket_prefix" validate:"required"`

	JobID string `json:"job_id" validate:"required"`
}

type PollWaypointDeploymentJobResponse struct{}

func (p PollWaypointDeploymentJobRequest) validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

func (a *Activities) PollWaypointDeploymentJob(
	ctx context.Context,
	req PollWaypointDeploymentJobRequest,
) (PollWaypointDeploymentJobResponse, error) {
	var resp PollWaypointDeploymentJobResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate waypoint deploy job: %w", err)
	}

	l := activity.GetLogger(ctx)

	client, err := a.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	logWriter := newLogEventWriter(l)
	fileWriter := newFileEventWriter()
	err = fileWriter.init()
	if err != nil {
		return resp, fmt.Errorf("unable to initialize job event tmp file for S3 upload: %w", err)
	}

	multiWriter := job.NewMultiWriter(logWriter, fileWriter)
	if err := job.Poll(ctx, client, req.JobID, multiWriter); err != nil {
		return resp, fmt.Errorf("unable to finish waypoint deployment job: %w", err)
	}

	// upload tmp file to S3 + cleanup
	uploadClient := uploader.NewS3Uploader(req.BucketName, req.BucketPrefix)
	if err := a.uploadEventFile(ctx, uploadClient, fileWriter); err != nil {
		return resp, fmt.Errorf("unable to upload events file to s3: %w", err)
	}

	return resp, nil
}

type fileEventWriter struct {
	fh      io.Writer
	fileLoc string
}

func newFileEventWriter() *fileEventWriter {
	return &fileEventWriter{}
}

func (f *fileEventWriter) init() error {
	// create a tmp file
	tmpFile, err := os.CreateTemp("", "instances-job-event")
	if err != nil {
		return err
	}
	f.fh = tmpFile
	f.fileLoc = tmpFile.Name()
	return nil
}

func (f fileEventWriter) Write(ev job.WaypointJobEvent) error {
	// convert event struct to json
	byts, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	// write each event on its own line in the file
	_, err = f.fh.Write(append(byts, []byte("\n")...))
	if err != nil {
		return err
	}

	return nil
}

type waypointDeploymentJobPollerImpl struct{}

type s3BlobUploader interface {
	UploadBlob(context.Context, []byte, string) error
}

type waypointDeploymentJobPoller interface {
	uploadEventFile(context.Context, s3BlobUploader, *fileEventWriter) error
}

var _ waypointDeploymentJobPoller = (*waypointDeploymentJobPollerImpl)(nil)

func (waypointDeploymentJobPollerImpl) uploadEventFile(ctx context.Context, client s3BlobUploader, fileWriter *fileEventWriter) error {
	contents, err := osReadFile(fileWriter.fileLoc)
	if err != nil {
		return fmt.Errorf("unable to read temp file: %s", err)
	}
	byts, err := json.Marshal(contents)
	if err != nil {
		return err
	}

	// upload file
	if err := client.UploadBlob(ctx, byts, eventFilename); err != nil {
		return fmt.Errorf("unable to upload events file to s3: %s", err)
	}

	// remove tmp file
	if err := osRemoveFile(fileWriter.fileLoc); err != nil {
		return fmt.Errorf("unable to remove temp file: %s", err)
	}

	return nil
}
