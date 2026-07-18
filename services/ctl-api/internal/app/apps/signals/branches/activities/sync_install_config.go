package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/installs"
)

type SyncInstallConfigInput struct {
	AppID               string          `json:"app_id" validate:"required"`
	InstallConfig       *config.Install `json:"install_config" validate:"required"`
	InstallConfigSyncID string          `json:"install_config_sync_id" validate:"required"`
	FilePath            string          `json:"file_path"`
}

type SyncInstallConfigOutput struct {
	InstallConfigVersionID string `json:"install_config_version_id"`
	InstallID              string `json:"install_id"`
	InstallName            string `json:"install_name"`
	Created                bool   `json:"created"`
	Changed                bool   `json:"changed"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) SyncInstallConfig(ctx context.Context, input *SyncInstallConfigInput) (*SyncInstallConfigOutput, error) {
	result, err := installs.SyncInstall(ctx, a.db, a.installHelpers, input.AppID, input.InstallConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to sync install config %s: %w", input.InstallConfig.Name, err)
	}

	if !result.Changed {
		return &SyncInstallConfigOutput{
			InstallID:   result.InstallID,
			InstallName: result.InstallName,
			Created:     false,
			Changed:     false,
		}, nil
	}

	version := app.InstallConfigVersion{
		InstallConfigSyncID: input.InstallConfigSyncID,
		InstallID:           result.InstallID,
		InstallName:         result.InstallName,
		FilePath:            input.FilePath,
		Created:             result.Created,
		Status:              app.NewCompositeStatus(ctx, app.StatusSuccess),
	}
	if err := a.db.WithContext(ctx).Create(&version).Error; err != nil {
		return nil, fmt.Errorf("unable to create install config version: %w", err)
	}

	if result.Diff != nil {
		if err := a.saveInstallConfigDiffBlob(ctx, version.ID, result.Diff); err != nil {
			a.l.Warn("unable to save install config diff blob", zap.Error(err))
		}
	}

	return &SyncInstallConfigOutput{
		InstallConfigVersionID: version.ID,
		InstallID:              result.InstallID,
		InstallName:            result.InstallName,
		Created:                result.Created,
		Changed:                true,
	}, nil
}

func (a *Activities) saveInstallConfigDiffBlob(ctx context.Context, installConfigVersionID string, diff interface{}) error {
	diffJSON, err := json.Marshal(diff)
	if err != nil {
		return fmt.Errorf("unable to marshal diff: %w", err)
	}

	blobID := domains.NewBlobID()
	s3Key := fmt.Sprintf("blobs/install_config_version_diffs/%s", blobID)

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
		Model(&app.InstallConfigVersion{}).
		Where(app.InstallConfigVersion{ID: installConfigVersionID}).
		Update("diff", string(metadataJSON))
	if res.Error != nil {
		return fmt.Errorf("unable to save diff: %w", res.Error)
	}

	return nil
}

// Ensure unused imports are consumed
var _ = installs.SyncInstall
