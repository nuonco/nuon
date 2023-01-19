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
	finishRequestFilename       string = "response.json"
	finishAssumeRoleSessionName string = "workers-orgs-finish"
)

type FinishRequest struct {
	MetadataBucket              string `validate:"required"`
	MetadataBucketAssumeRoleARN string `validate:"required"`
	MetadataBucketPrefix        string `validate:"required"`

	Response       *sharedv1.ResponseRef `faker:"-"`
	ResponseStatus sharedv1.ResponseStatus
	ErrorMessage   string
}

func (s FinishRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

type FinishResponse struct{}

func NewFinishActivity() *finishActivity {
	return &finishActivity{
		finisher: &finisherImpl{},
	}
}

type finishActivity struct {
	finisher finisher
}

func (a *finishActivity) FinishRequest(ctx context.Context, req *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
	resp := &sharedv1.FinishActivityResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	// create upload client
	assumeRoleOpt := uploader.WithAssumeRoleARN(req.MetadataBucketAssumeRoleArn)
	assumeRoleSessionOpt := uploader.WithAssumeSessionName(startAssumeRoleSessionName)
	uploadClient := uploader.NewS3Uploader(req.MetadataBucket, req.MetadataBucketPrefix,
		assumeRoleOpt, assumeRoleSessionOpt)

	obj := a.finisher.getResponse(req)
	if err := a.finisher.writeRequestFile(ctx, uploadClient, obj); err != nil {
		return resp, fmt.Errorf("unable to write request: %w", err)
	}

	return resp, nil
}

type finisher interface {
	getResponse(*sharedv1.FinishActivityRequest) *sharedv1.Response
	writeRequestFile(context.Context, finisherUploadClient, *sharedv1.Response) error
}

type finisherImpl struct{}

var _ finisher = (*finisherImpl)(nil)

func (s *finisherImpl) getResponse(req *sharedv1.FinishActivityRequest) *sharedv1.Response {
	return &sharedv1.Response{
		Status:   req.Status,
		Response: req.ResponseRef,
	}
}

type finisherUploadClient interface {
	UploadBlob(context.Context, []byte, string) error
}

func (s *finisherImpl) writeRequestFile(ctx context.Context, client finisherUploadClient, req *sharedv1.Response) error {
	byts, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("unable to convert request to json: %w", err)
	}

	if err := client.UploadBlob(ctx, byts, finishRequestFilename); err != nil {
		return fmt.Errorf("unable to finish uploading response: %w", err)
	}

	return nil
}
