package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/aws/s3uploader"
)

type UploadAWSCloudFormationStackVersionTemplateRequest struct {
	BucketKey string `validate:"required"`
	Template  []byte `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UploadAWSCloudFormationStackVersionTemplate(ctx context.Context, req *UploadAWSCloudFormationStackVersionTemplateRequest) error {
	uploader, err := s3uploader.NewS3Uploader(a.v,
		s3uploader.WithBucketName(a.cfg.AWSCloudFormationStackTemplateBucket),
		s3uploader.WithCredentials(&credentials.Config{
			Region:     "us-east-1",
			UseDefault: true,
		}),
	)
	if err != nil {
		return errors.Wrap(err, "unable to create s3 uploader")
	}

	if err := uploader.UploadBlob(ctx, req.Template, req.BucketKey); err != nil {
		return errors.Wrap(err, "unable to upload template")
	}

	return nil
}
