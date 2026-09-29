package cloudformation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type BlobUploader interface {
	UploadBlob(ctx context.Context, blob []byte, key string) error
}

func UploadCustomNestedStackTemplates(ctx context.Context, uploader BlobUploader, baseURL string, stackConfig *app.AppStackConfig) error {
	sourceURL := func(contentsHash, templateURL string) string {
		if baseURL == "" {
			return ""
		}
		return CustomNestedStackTemplateURL(baseURL, stackConfig.OrgID, stackConfig.AppID, contentsHash, templateURL)
	}

	for i, stack := range stackConfig.CustomNestedStacks {
		if stack.Contents == "" {
			if stack.ContentsHash != "" {
				stackConfig.CustomNestedStacks[i].Status = config.CustomNestedStackStatusReady
				stackConfig.CustomNestedStacks[i].TemplateSourceURL = sourceURL(stack.ContentsHash, stack.TemplateURL)
			}
			continue
		}

		hash := sha256.Sum256([]byte(stack.Contents))
		contentsHash := hex.EncodeToString(hash[:])

		s3Key := CustomNestedStackS3Key(stackConfig.OrgID, stackConfig.AppID, contentsHash, stack.TemplateURL)

		if err := uploader.UploadBlob(ctx, []byte(stack.Contents), s3Key); err != nil {
			stackConfig.CustomNestedStacks[i].Status = config.CustomNestedStackStatusError
			return fmt.Errorf("unable to upload custom nested stack template %q: %w", stack.Name, err)
		}

		stackConfig.CustomNestedStacks[i].ContentsHash = contentsHash
		stackConfig.CustomNestedStacks[i].Contents = ""
		stackConfig.CustomNestedStacks[i].Status = config.CustomNestedStackStatusReady
		stackConfig.CustomNestedStacks[i].TemplateSourceURL = sourceURL(contentsHash, stack.TemplateURL)
	}

	return nil
}
