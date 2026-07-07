package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type CreateInstallAppConfigVersionInput struct {
	InstallID      string                 `json:"install_id" validate:"required"`
	OldAppConfigID string                 `json:"old_app_config_id"`
	NewAppConfigID string                 `json:"new_app_config_id" validate:"required"`
	Diff           *app.InstallConfigDiff `json:"diff,omitempty"`
	Metadata       map[string]string      `json:"metadata,omitempty"`
	AppBranchRunID string                 `json:"app_branch_run_id,omitempty"`
	InstallGroupID string                 `json:"install_group_id,omitempty"`
}

type CreateInstallAppConfigVersionOutput struct {
	ID string `json:"id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) CreateInstallAppConfigVersion(ctx context.Context, input *CreateInstallAppConfigVersionInput) (*CreateInstallAppConfigVersionOutput, error) {
	version := app.InstallAppConfigVersion{
		InstallID:      input.InstallID,
		OldAppConfigID: input.OldAppConfigID,
		NewAppConfigID: input.NewAppConfigID,
		Metadata:       input.Metadata,
		Status:         app.NewCompositeStatus(ctx, app.StatusSuccess),
	}
	if input.AppBranchRunID != "" {
		version.AppBranchRunID = &input.AppBranchRunID
	}
	if input.InstallGroupID != "" {
		version.InstallGroupID = &input.InstallGroupID
	}
	if err := a.db.WithContext(ctx).Create(&version).Error; err != nil {
		return nil, fmt.Errorf("unable to create install app config version: %w", err)
	}

	if input.Diff != nil {
		if err := a.saveDiffBlob(ctx, version.ID, input.Diff); err != nil {
			a.l.Warn("unable to save config diff blob", zap.Error(err))
		}
	}

	return &CreateInstallAppConfigVersionOutput{ID: version.ID}, nil
}

func (a *Activities) saveDiffBlob(ctx context.Context, installConfigVersionID string, diff *app.InstallConfigDiff) error {
	diffJSON, err := json.Marshal(diff)
	if err != nil {
		return fmt.Errorf("unable to marshal diff: %w", err)
	}

	blobID := domains.NewBlobID()
	s3Key := fmt.Sprintf("blobs/install_config_diffs/%s", blobID)

	reader := strings.NewReader(string(diffJSON))
	checksum, err := a.blobSvc.UploadStream(ctx, s3Key, reader)
	if err != nil {
		return fmt.Errorf("unable to upload diff to S3: %w", err)
	}

	metadata := blobstore.BlobMetadata{
		BlobID:      blobID,
		S3Key:       s3Key,
		Size:        int64(len(diffJSON)),
		ContentType: "application/json",
		Checksum:    checksum,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("unable to marshal blob metadata: %w", err)
	}

	res := a.db.WithContext(ctx).
		Model(&app.InstallAppConfigVersion{}).
		Where(app.InstallAppConfigVersion{ID: installConfigVersionID}).
		Update("diff", string(metadataJSON))
	if res.Error != nil {
		return fmt.Errorf("unable to save diff: %w", res.Error)
	}

	return nil
}
