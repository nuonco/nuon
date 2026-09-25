package activities

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/aws/s3uploader"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

type UploadCustomNestedStackTemplatesRequest struct {
	AppStackConfigID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UploadCustomNestedStackTemplates(ctx context.Context, req *UploadCustomNestedStackTemplatesRequest) error {
	var stackConfig app.AppStackConfig
	res := a.db.WithContext(ctx).First(&stackConfig, "id = ?", req.AppStackConfigID)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return temporal.NewNonRetryableApplicationError("not found", "not found", res.Error, "")
		}
		return fmt.Errorf("unable to get app stack config: %w", res.Error)
	}

	uploader, err := s3uploader.NewS3Uploader(a.v,
		s3uploader.WithBucketName(a.cfg.AWSCloudFormationStackTemplateBucket),
		s3uploader.WithCredentials(a.cfg.CFTemplateUploadCreds()),
	)
	if err != nil {
		return fmt.Errorf("unable to create s3 uploader: %w", err)
	}

	if err := cloudformation.UploadCustomNestedStackTemplates(ctx, uploader, a.cfg.AWSCloudFormationStackTemplateBaseURL, &stackConfig); err != nil {
		return err
	}

	res = a.db.WithContext(ctx).
		Model(&stackConfig).
		Select("custom_nested_stacks").
		Updates(&stackConfig)
	if res.Error != nil {
		return fmt.Errorf("unable to update app stack config: %w", res.Error)
	}

	return nil
}
