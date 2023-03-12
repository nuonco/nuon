package meta

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	"github.com/powertoolsdev/mono/pkg/uploader"
	"google.golang.org/protobuf/proto"
)

const (
	startRequestFilename       string = "request.json"
	startAssumeRoleSessionName string = "workers-orgs-start"
)

type WorkflowInfo struct {
	ID string `validate:"required"`
}

func NewStartActivity() *startActivity {
	v := validator.New()
	return &startActivity{
		v:       v,
		starter: &starterImpl{},
	}
}

type startActivity struct {
	v       *validator.Validate
	starter starter
}

func (s *startActivity) StartRequest(ctx context.Context, req *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
	resp := &sharedv1.StartActivityResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	// create upload client
	uploadClient, err := uploader.NewS3Uploader(s.v,
		uploader.WithBucketName(req.MetadataBucket),
		uploader.WithPrefix(req.MetadataBucketPrefix),
		uploader.WithAssumeSessionName(req.MetadataBucket),
		uploader.WithAssumeRoleARN(req.MetadataBucketAssumeRoleArn))

	if err != nil {
		return nil, fmt.Errorf("unable to get uploader: %w", err)
	}

	obj := s.starter.getRequest(req)
	if err := s.starter.writeRequestFile(ctx, uploadClient, obj); err != nil {
		return resp, fmt.Errorf("unable to write request: %w", err)
	}

	return resp, nil
}

type starter interface {
	getRequest(*sharedv1.StartActivityRequest) *sharedv1.Request
	writeRequestFile(context.Context, starterUploadClient, *sharedv1.Request) error
}

type starterImpl struct{}

var _ starter = (*starterImpl)(nil)

func (s *starterImpl) getRequest(req *sharedv1.StartActivityRequest) *sharedv1.Request {
	return &sharedv1.Request{
		WorkflowId: req.WorkflowInfo.Id,
		// TODO: parse temporal memo and map to our own types
		Request: req.RequestRef,
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
