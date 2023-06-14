package meta

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/aws/s3uploader"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	"google.golang.org/protobuf/proto"
)

const (
	finishRequestFilename       string = "response.json"
	finishAssumeRoleSessionName string = "workers-orgs-finish"
)

func NewFinishActivity() *finishActivity {
	v := validator.New()
	return &finishActivity{
		v:        v,
		finisher: &finisherImpl{},
	}
}

type finishActivity struct {
	v        *validator.Validate
	finisher finisher
}

func (a *finishActivity) FinishRequest(ctx context.Context, req *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
	resp := &sharedv1.FinishActivityResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	// create upload client
	uploadClient, err := s3uploader.NewS3Uploader(a.v,
		s3uploader.WithBucketName(req.MetadataBucket),
		s3uploader.WithPrefix(req.MetadataBucketPrefix),
		s3uploader.WithAssumeSessionName(req.MetadataBucket),
		s3uploader.WithAssumeRoleARN(req.MetadataBucketAssumeRoleArn))

	if err != nil {
		return nil, fmt.Errorf("unable to get uploader: %w", err)
	}
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
