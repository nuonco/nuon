package build

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-waypoint/job"
	"go.temporal.io/sdk/activity"
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

	// TODO: read tmp file and upload it to S3
	if err := os.Remove(fileWriter.fileLoc); err != nil {
		return resp, fmt.Errorf("unable to remove local tmp file: %w", err)
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
	tmpFile, err := os.CreateTemp("", "deployments-job-event")
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
