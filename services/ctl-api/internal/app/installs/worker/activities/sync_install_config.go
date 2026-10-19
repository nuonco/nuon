package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/installs"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

type SyncInstallConfigInput struct {
	AppID               string          `json:"app_id" validate:"required"`
	InstallConfig       *config.Install `json:"install_config" validate:"required"`
	InstallConfigSyncID string          `json:"install_config_sync_id" validate:"required"`
	FilePath            string          `json:"file_path"`
	CreateOnly          bool            `json:"create_only,omitempty"`
}

type SyncInstallConfigOutput struct {
	InstallConfigVersionID string `json:"install_config_version_id"`
	InstallID              string `json:"install_id"`
	InstallName            string `json:"install_name"`
	Created                bool   `json:"created"`
	Changed                bool   `json:"changed"`
	AppBranchChanged       bool   `json:"app_branch_changed,omitempty"`
	AppBranchID            string `json:"app_branch_id,omitempty"`
	WorkflowID             string `json:"workflow_id,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) SyncInstallConfig(ctx context.Context, input *SyncInstallConfigInput) (*SyncInstallConfigOutput, error) {
	hash, err := request.Hash(input)
	if err != nil {
		return nil, err
	}
	key := app.InstallConfigVersion{InstallConfigSyncID: input.InstallConfigSyncID, InstallName: input.InstallConfig.Name}
	var version app.InstallConfigVersion
	if err := a.db.WithContext(ctx).Where(key).Limit(1).Find(&version).Error; err != nil {
		return nil, fmt.Errorf("unable to load install config version: %w", err)
	}
	if version.ID != "" && version.Metadata["request_hash"] != "" && version.Metadata["request_hash"] != hash {
		return nil, fmt.Errorf("install config sync was already used with a different config")
	}
	if version.ID != "" && version.Status.Status == app.StatusSuccess {
		return &SyncInstallConfigOutput{
			InstallConfigVersionID: version.ID,
			InstallID:              version.InstallID, InstallName: version.InstallName, Created: version.Created, Changed: true,
			AppBranchChanged: version.Metadata["app_branch_changed"] == "true",
			AppBranchID:      version.Metadata["app_branch_id"], WorkflowID: version.Metadata["workflow_id"],
		}, nil
	}
	if input.CreateOnly && version.ID == "" {
		var count int64
		if err := a.db.WithContext(ctx).Model(&app.Install{}).
			Where(app.Install{AppID: input.AppID, Name: input.InstallConfig.Name}).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("proposed install %q already exists; sync it through its install config workflow", input.InstallConfig.Name)
		}
	}
	result, err := installs.SyncInstall(ctx, a.db, a.helpers, input.AppID, input.InstallConfig, func(db *gorm.DB, planned *sync.InstallSyncResult) error {
		if version.ID != "" {
			return nil
		}
		pending := app.InstallConfigVersion{
			InstallConfigSyncID: input.InstallConfigSyncID,
			InstallID:           planned.InstallID, InstallName: planned.InstallName, Created: planned.Created,
			FilePath: input.FilePath,
			Status:   app.NewCompositeStatus(ctx, app.StatusPending),
			Metadata: map[string]string{
				"request_hash": hash, "app_branch_id": planned.AppBranchID,
				"app_branch_changed": strconv.FormatBool(planned.AppBranchChanged),
			},
		}
		if err := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&pending).Error; err != nil {
			return fmt.Errorf("unable to record install config change: %w", err)
		}
		if err := db.WithContext(ctx).Where(key).First(&version).Error; err != nil {
			return err
		}
		if version.Metadata["request_hash"] != hash {
			return fmt.Errorf("install config sync was already used with a different config")
		}
		if version.InstallID != planned.InstallID {
			return fmt.Errorf("install config change was already recorded for install %s", version.InstallID)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to sync install config %s: %w", input.InstallConfig.Name, err)
	}

	if version.ID == "" {
		return &SyncInstallConfigOutput{
			InstallID:   result.InstallID,
			InstallName: result.InstallName,
			Created:     false,
			Changed:     false,
		}, nil
	}

	if version.Created {
		if err := a.helpers.EnsureInstallQueues(ctx, version.InstallID); err != nil {
			return nil, err
		}
		if err := a.helpers.EnsureInstallRunner(ctx, version.InstallID); err != nil {
			return nil, err
		}
		if err := a.helpers.ReconcileInstallComponents(ctx, version.InstallID); err != nil {
			return nil, err
		}
		if err := a.helpers.ReconcileInstallActions(ctx, version.InstallID); err != nil {
			return nil, err
		}
		if err := a.helpers.ReconcileInstallRunbooks(ctx, version.InstallID); err != nil {
			return nil, err
		}
	}
	metadata := version.Metadata
	metadata["app_branch_changed"] = strconv.FormatBool(result.AppBranchChanged || metadata["app_branch_changed"] == "true")
	if result.AppBranchID != "" {
		metadata["app_branch_id"] = result.AppBranchID
	}
	complete := func(db *gorm.DB, workflowID string) error {
		metadata["workflow_id"] = workflowID
		return db.WithContext(ctx).Model(&app.InstallConfigVersion{}).Where(app.InstallConfigVersion{ID: version.ID}).
			Updates(map[string]any{"status": app.NewCompositeStatus(ctx, app.StatusSuccess), "metadata": metadata}).Error
	}
	workflowID := ""
	if input.InstallConfig.CloudConnection != "" {
		workflowType := app.WorkflowTypeReprovisionStack
		workflowMetadata := map[string]string{"install_config_version_id": version.ID}
		requestID := version.ID
		workflowHash := hash
		if version.Created {
			workflowType = app.WorkflowTypeProvision
			workflowMetadata[app.WorkflowMetadataKeyStackOnly] = "true"
			requestID = "git-install-initial:" + version.InstallID
			workflowHash, err = request.Hash(version.InstallID)
			if err != nil {
				return nil, err
			}
		}
		wf, _, err := a.helpers.RunIdempotentInstallWorkflow(ctx, helpers.IdempotentInstallWorkflowRequest{
			InstallID: version.InstallID, WorkflowType: workflowType, Metadata: workflowMetadata,
			RequestID: requestID, RequestHash: workflowHash, Operation: "install-config-stack",
			QueueName: queuenames.InstallWorkflowsQueueName,
		}, &helpers.IdempotentInstallWorkflowHooks{
			AfterCreate: func(tx *gorm.DB, wf *app.Workflow) error { return complete(tx, wf.ID) },
		})
		if err != nil {
			return nil, fmt.Errorf("unable to schedule install stack: %w", err)
		}
		workflowID = wf.ID
	}
	if err := complete(a.db, workflowID); err != nil {
		return nil, fmt.Errorf("unable to complete install config version: %w", err)
	}

	if result.Diff != nil {
		if err := a.saveInstallConfigDiffBlob(ctx, version.ID, result.Diff); err != nil {
			a.l.Warn("unable to save install config diff blob", zap.Error(err))
		}
	}

	return &SyncInstallConfigOutput{
		InstallConfigVersionID: version.ID,
		InstallID:              version.InstallID,
		InstallName:            version.InstallName,
		Created:                version.Created,
		Changed:                true,
		AppBranchChanged:       metadata["app_branch_changed"] == "true",
		AppBranchID:            metadata["app_branch_id"],
		WorkflowID:             workflowID,
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
