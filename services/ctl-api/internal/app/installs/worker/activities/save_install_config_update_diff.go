package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type SaveInstallConfigUpdateDiffInput struct {
	InstallConfigUpdateID string `json:"install_config_update_id" validate:"required"`
	DiffJSON              string `json:"diff_json" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) SaveInstallConfigUpdateDiff(ctx context.Context, input *SaveInstallConfigUpdateDiffInput) error {
	if err := a.v.Struct(input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	// Upload diff to S3
	blobID := domains.NewBlobID()
	s3Key := fmt.Sprintf("blobs/install_config_diffs/%s", blobID)

	reader := strings.NewReader(input.DiffJSON)
	checksum, err := a.blobSvc.UploadStream(ctx, s3Key, reader)
	if err != nil {
		return fmt.Errorf("unable to upload diff to S3: %w", err)
	}

	// Build blob metadata JSONB
	metadata := blobstore.BlobMetadata{
		BlobID:      blobID,
		S3Key:       s3Key,
		Size:        int64(len(input.DiffJSON)),
		ContentType: "application/json",
		Checksum:    checksum,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("unable to marshal blob metadata: %w", err)
	}

	// Update the InstallConfigUpdate record with the diff metadata
	res := a.db.WithContext(ctx).
		Model(&app.InstallConfigUpdate{}).
		Where(app.InstallConfigUpdate{ID: input.InstallConfigUpdateID}).
		Update("diff", string(metadataJSON))
	if res.Error != nil {
		return fmt.Errorf("unable to save diff on install config update %s: %w", input.InstallConfigUpdateID, res.Error)
	}

	return nil
}
