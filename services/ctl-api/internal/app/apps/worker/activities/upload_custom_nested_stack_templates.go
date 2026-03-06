package activities

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/aws/s3uploader"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"
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
		s3uploader.WithCredentials(&credentials.Config{
			Region:     a.cfg.AWSCloudFormationStackTemplateBucketRegion,
			UseDefault: true,
		}),
	)
	if err != nil {
		return fmt.Errorf("unable to create s3 uploader: %w", err)
	}

	for i, stack := range stackConfig.CustomNestedStacks {
		if stack.Contents == "" {
			continue
		}

		hash := sha256.Sum256([]byte(stack.Contents))
		contentsHash := hex.EncodeToString(hash[:])

		s3Key := cloudformation.CustomNestedStackS3Key(stackConfig.OrgID, stackConfig.AppID, contentsHash, stack.TemplateURL)

		if err := uploader.UploadBlob(ctx, []byte(stack.Contents), s3Key); err != nil {
			return fmt.Errorf("unable to upload custom nested stack template %q: %w", stack.Name, err)
		}

		stackConfig.CustomNestedStacks[i].ContentsHash = contentsHash
		stackConfig.CustomNestedStacks[i].Contents = ""
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
