package meta

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-uploader"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
	"google.golang.org/protobuf/proto"
)

const (
	startRequestFilename       string = "request.json"
	startAssumeRoleSessionName string = "workers-orgs-start"
)

type WorkflowInfo struct {
	ID string `validate:"required"`
}

type StartRequest struct {
	MetadataBucket              string `validate:"required"`
	MetadataBucketAssumeRoleARN string `validate:"required"`
	MetadataBucketPrefix        string `validate:"required"`

	Request      *sharedv1.RequestRef `validate:"required" faker:"-"`
	WorkflowInfo WorkflowInfo         `validate:"required"`
}

type StartResponse struct{}

func (s StartRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

func NewStartActivity() *startActivity {
	return &startActivity{
		starter: &starterImpl{},
	}
}

type startActivity struct {
	starter starter
}

func (s *startActivity) StartRequest(ctx context.Context, req StartRequest) (StartResponse, error) {
	var resp StartResponse

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	// create upload client
	assumeRoleOpt := uploader.WithAssumeRoleARN(req.MetadataBucketAssumeRoleARN)
	assumeRoleSessionOpt := uploader.WithAssumeSessionName(startAssumeRoleSessionName)
	uploadClient := uploader.NewS3Uploader(req.MetadataBucket, req.MetadataBucketPrefix,
		assumeRoleOpt, assumeRoleSessionOpt)

	obj := s.starter.getRequest(req)
	if err := s.starter.writeRequestFile(ctx, uploadClient, obj); err != nil {
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
		Request: req.Request,
	}
}

type starterUploadClient interface {
	UploadBlob(context.Context, []byte, string) error
}

func (s *starterImpl) writeRequestFile(ctx context.Context, client starterUploadClient, req *sharedv1.Request) error {
	byts, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("unable to convert request to json: %w", err)
	}

	if err := client.UploadBlob(ctx, byts, startRequestFilename); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}
