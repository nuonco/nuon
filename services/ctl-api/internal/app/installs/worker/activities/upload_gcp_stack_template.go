package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/gcp/gcsuploader"
)

type UploadGCPStackTemplateRequest struct {
	BucketKey string `validate:"required"`
	Template  []byte `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UploadGCPStackTemplate(ctx context.Context, req *UploadGCPStackTemplateRequest) error {
	if a.cfg.GCPStackTemplateBucket == "" {
		return nil
	}

	uploader := gcsuploader.New(a.cfg.GCPStackTemplateBucket)
	if err := uploader.Upload(ctx, req.Template, req.BucketKey); err != nil {
		return errors.Wrap(err, "unable to upload GCP template to GCS")
	}

	return nil
}
