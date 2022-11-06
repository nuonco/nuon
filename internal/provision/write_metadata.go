package provision

import (
	"context"
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-uploader"
)

type UploadMetadataRequest struct {
	Info workflow.Info

	BucketName   string `json:"bucket_name"   validate:"required"`
	BucketPrefix string `json:"bucket_prefix" validate:"required"`
}

func (p UploadMetadataRequest) validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type UploadResultResponse struct{}

type metadataUploaderImpl struct{}

type metadataUploader interface {
	uploadMetadata(context.Context, UploadMetadataRequest) error
}

var _ metadataUploader = (*metadataUploaderImpl)(nil)

func (a *Activities) UploadMetadata(
	ctx context.Context,
	req UploadMetadataRequest,
) (*UploadResultResponse, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	err := a.uploadMetadata(ctx, req)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (m *metadataUploaderImpl) uploadMetadata(ctx context.Context, req UploadMetadataRequest) error {
	s3upload := uploader.NewS3Uploader(req.BucketName, req.BucketPrefix)

	bytes, err := json.Marshal(req.Info)
	if err != nil {
		return err
	}

	err = s3upload.UploadBlob(ctx, bytes, "workflow.json")
	if err != nil {
		return err
	}

	return nil
}
