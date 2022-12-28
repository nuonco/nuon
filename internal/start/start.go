package start

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-uploader"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

const (
	startRequestFilename       string = "request.json"
	startAssumeRoleSessionName string = "workers-deployments-start"
)

type WorkflowInfo struct {
	ID string `validate:"required"`
}

type StartRequest struct {
	DeploymentsBucket              string `validate:"required"`
	DeploymentsBucketAssumeRoleARN string `validate:"required"`
	DeploymentsBucketPrefix        string `validate:"required"`

	Request      *deploymentsv1.StartRequest `validate:"required"`
	WorkflowInfo WorkflowInfo                `validate:"required"`
}

func (s StartRequest) validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

type StartResponse struct{}

func (a *Activities) StartRequest(ctx context.Context, req StartRequest) (StartResponse, error) {
	var resp StartResponse

	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	// create upload client
	assumeRoleOpt := uploader.WithAssumeRoleARN(req.DeploymentsBucketAssumeRoleARN)
	assumeRoleSessionOpt := uploader.WithAssumeSessionName(startAssumeRoleSessionName)
	uploadClient := uploader.NewS3Uploader(req.DeploymentsBucket, req.DeploymentsBucketPrefix,
		assumeRoleOpt, assumeRoleSessionOpt)

	obj := a.starter.getRequest(req)
	if err := a.starter.writeRequestFile(ctx, uploadClient, obj); err != nil {
		return resp, fmt.Errorf("unable to write request: %w", err)
	}

	return resp, nil
}

type starter interface {
	getRequest(StartRequest) *sharedv1.Request
	writeRequestFile(context.Context, starterUploadClient, *sharedv1.Request) error
}

type starterImpl struct{}

var _ starter = (*starterImpl)(nil)

func (s *starterImpl) getRequest(req StartRequest) *sharedv1.Request {
	return &sharedv1.Request{
		WorkflowId: req.WorkflowInfo.ID,
		// TODO: parse temporal memo and map to our own types
		Request: &sharedv1.RequestRef{
			Request: &sharedv1.RequestRef_DeploymentStart{
				DeploymentStart: req.Request,
			},
		},
	}
}

type starterUploadClient interface {
	UploadBlob(context.Context, []byte, string) error
}

func (s *starterImpl) writeRequestFile(ctx context.Context, client starterUploadClient, req *sharedv1.Request) error {
	byts, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("unable to convert request to json: %w", err)
	}

	if err := client.UploadBlob(ctx, byts, startRequestFilename); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}
