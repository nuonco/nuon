package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/aws/s3uploader"
)

type UploadAWSCloudFormationStackVersionTemplateRequest struct {
	BucketKey string `validate:"required"`
	Template  []byte `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UploadAWSCloudFormationStackVersionTemplate(ctx context.Context, req *UploadAWSCloudFormationStackVersionTemplateRequest) error {
	creds := &credentials.Config{
		Region:     a.cfg.AWSCloudFormationStackTemplateBucketRegion,
		UseDefault: true,
	}
	if a.cfg.UseLocal {
		creds.S3Endpoint = a.cfg.LocalS3Endpoint
		creds.S3ForcePathStyle = true
	}

	uploader, err := s3uploader.NewS3Uploader(a.v,
		s3uploader.WithBucketName(a.cfg.AWSCloudFormationStackTemplateBucket),
		s3uploader.WithCredentials(creds),
	)
	if err != nil {
		return errors.Wrap(err, "unable to create s3 uploader")
	}

	if err := uploader.UploadBlob(ctx, req.Template, req.BucketKey); err != nil {
		return errors.Wrap(err, "unable to upload template")
	}

	return nil
}
