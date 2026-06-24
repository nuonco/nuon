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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type CreateInstallConfigUpdateWorkflowInput struct {
	InstallID      string       `json:"install_id"`
	NewAppConfigID string       `json:"new_app_config_id"`
	AppBranchRunID string       `json:"app_branch_run_id"`
	InstallGroupID string       `json:"install_group_id"`
	PlanOnly       bool         `json:"plan_only"`
	Callback       callback.Ref `json:"callback,omitempty"`
}

type CreateInstallConfigUpdateWorkflowOutput struct {
	WorkflowID            string `json:"workflow_id"`
	InstallConfigUpdateID string `json:"install_config_update_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallConfigUpdateWorkflow(ctx context.Context, input *CreateInstallConfigUpdateWorkflowInput) (*CreateInstallConfigUpdateWorkflowOutput, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).First(&install, "id = ?", input.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	diff, err := a.computeInstallConfigDiff(ctx, install.AppConfigID, input.NewAppConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to compute config diff: %w", err)
	}

	update := app.InstallConfigUpdate{
		AppBranchRunID: input.AppBranchRunID,
		InstallGroupID: input.InstallGroupID,
		InstallID:      input.InstallID,
		OldAppConfigID: install.AppConfigID,
		NewAppConfigID: input.NewAppConfigID,
		Status:         app.NewCompositeStatus(ctx, app.StatusPending),
	}
	if err := a.db.WithContext(ctx).Create(&update).Error; err != nil {
		return nil, fmt.Errorf("unable to create install config update: %w", err)
	}

	if err := a.saveDiffBlob(ctx, update.ID, diff); err != nil {
		a.l.Warn("unable to save config diff blob", zap.Error(err))
	}

	metadata := map[string]string{
		"new_app_config_id":        input.NewAppConfigID,
		"app_branch_run_id":        input.AppBranchRunID,
		"install_group_id":         input.InstallGroupID,
		"install_config_update_id": update.ID,
	}

	wf, err := a.installHelpers.CreateWorkflow(
		ctx,
		input.InstallID,
		app.WorkflowTypeAppBranchConfigUpdate,
		metadata,
		input.PlanOnly,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create install config update workflow: %w", err)
	}

	if err := a.db.WithContext(ctx).
		Model(&update).
		Update("workflow_id", wf.ID).Error; err != nil {
		return nil, fmt.Errorf("unable to link workflow to install config update: %w", err)
	}

	queue, err := a.queueClient.GetQueueByOwner(ctx, input.InstallID, "installs")
	if err != nil {
		return nil, fmt.Errorf("unable to find queue for install %s: %w", input.InstallID, err)
	}

	if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:  queue.ID,
		Signal:   &executeflow.Signal{WorkflowID: wf.ID},
		Callback: input.Callback,
	}); err != nil {
		return nil, fmt.Errorf("unable to enqueue workflow for install %s: %w", input.InstallID, err)
	}

	return &CreateInstallConfigUpdateWorkflowOutput{
		WorkflowID:            wf.ID,
		InstallConfigUpdateID: update.ID,
	}, nil
}

func (a *Activities) computeInstallConfigDiff(ctx context.Context, oldAppConfigID, newAppConfigID string) (*app.InstallConfigDiff, error) {
	var newAppCfg app.AppConfig
	if err := a.db.WithContext(ctx).
		Preload("ComponentConfigConnections").
		Preload("SandboxConfig").
		Preload("StackConfig").
		First(&newAppCfg, "id = ?", newAppConfigID).Error; err != nil {
		return nil, fmt.Errorf("unable to get new app config: %w", err)
	}

	diff := &app.InstallConfigDiff{
		Added:     []app.ComponentDiffEntry{},
		Removed:   []app.ComponentDiffEntry{},
		Changed:   []app.ComponentDiffEntry{},
		Unchanged: []app.ComponentDiffEntry{},
	}

	newConnByComponent := make(map[string]*app.ComponentConfigConnection, len(newAppCfg.ComponentConfigConnections))
	for i := range newAppCfg.ComponentConfigConnections {
		ccc := &newAppCfg.ComponentConfigConnections[i]
		newConnByComponent[ccc.ComponentID] = ccc
	}

	if oldAppConfigID != "" {
		var oldAppCfg app.AppConfig
		if err := a.db.WithContext(ctx).
			Preload("ComponentConfigConnections").
			Preload("SandboxConfig").
			Preload("StackConfig").
			First(&oldAppCfg, "id = ?", oldAppConfigID).Error; err == nil {

			oldConnByComponent := make(map[string]*app.ComponentConfigConnection, len(oldAppCfg.ComponentConfigConnections))
			for i := range oldAppCfg.ComponentConfigConnections {
				ccc := &oldAppCfg.ComponentConfigConnections[i]
				oldConnByComponent[ccc.ComponentID] = ccc
			}

			for componentID, oldConn := range oldConnByComponent {
				newConn, exists := newConnByComponent[componentID]
				if !exists {
					diff.Removed = append(diff.Removed, app.ComponentDiffEntry{
						ComponentID:   componentID,
						ComponentName: oldConn.ComponentName,
						ComponentType: string(oldConn.Type),
						OldChecksum:   oldConn.Checksum,
					})
					continue
				}

				if oldConn.Checksum != "" && newConn.Checksum != "" && oldConn.Checksum == newConn.Checksum {
					diff.Unchanged = append(diff.Unchanged, app.ComponentDiffEntry{
						ComponentID:   componentID,
						ComponentName: newConn.ComponentName,
						ComponentType: string(newConn.Type),
						OldChecksum:   oldConn.Checksum,
						NewChecksum:   newConn.Checksum,
					})
				} else {
					diff.Changed = append(diff.Changed, app.ComponentDiffEntry{
						ComponentID:   componentID,
						ComponentName: newConn.ComponentName,
						ComponentType: string(newConn.Type),
						OldChecksum:   oldConn.Checksum,
						NewChecksum:   newConn.Checksum,
					})
				}

				delete(newConnByComponent, componentID)
			}

			for componentID, newConn := range newConnByComponent {
				diff.Added = append(diff.Added, app.ComponentDiffEntry{
					ComponentID:   componentID,
					ComponentName: newConn.ComponentName,
					ComponentType: string(newConn.Type),
					NewChecksum:   newConn.Checksum,
				})
			}

			if oldAppCfg.SandboxConfig.ID != newAppCfg.SandboxConfig.ID {
				diff.SandboxChanged = true
				diff.SandboxOldID = oldAppCfg.SandboxConfig.ID
				diff.SandboxNewID = newAppCfg.SandboxConfig.ID
			}

			if oldAppCfg.StackConfig.ID != newAppCfg.StackConfig.ID {
				diff.StackChanged = true
				diff.StackOldID = oldAppCfg.StackConfig.ID
				diff.StackNewID = newAppCfg.StackConfig.ID
			}
		}
	}

	if oldAppConfigID == "" {
		for componentID, newConn := range newConnByComponent {
			diff.Added = append(diff.Added, app.ComponentDiffEntry{
				ComponentID:   componentID,
				ComponentName: newConn.ComponentName,
				ComponentType: string(newConn.Type),
				NewChecksum:   newConn.Checksum,
			})
		}

		if newAppCfg.SandboxConfig.ID != "" {
			diff.SandboxChanged = true
			diff.SandboxNewID = newAppCfg.SandboxConfig.ID
		}

		if newAppCfg.StackConfig.ID != "" {
			diff.StackChanged = true
			diff.StackNewID = newAppCfg.StackConfig.ID
		}
	}

	return diff, nil
}

func (a *Activities) saveDiffBlob(ctx context.Context, installConfigUpdateID string, diff *app.InstallConfigDiff) error {
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
		Model(&app.InstallConfigUpdate{}).
		Where(app.InstallConfigUpdate{ID: installConfigUpdateID}).
		Update("diff", string(metadataJSON))
	if res.Error != nil {
		return fmt.Errorf("unable to save diff: %w", res.Error)
	}

	return nil
}
