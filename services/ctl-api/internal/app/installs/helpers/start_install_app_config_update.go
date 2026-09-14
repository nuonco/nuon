package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type StartInstallAppConfigUpdateInput struct {
	InstallID                string
	NewAppConfigID           string
	AppBranchRunID           string
	InstallGroupID           string
	AppReleaseID             string
	OperatingModelID         string
	ReleaseComponentBuildIDs map[string]string
	ReleaseSandboxBuildID    string
	Metadata                 map[string]string
	PlanOnly                 bool
	ApprovalOption           app.InstallApprovalOption
	Callback                 callback.Ref
}

func (h *Helpers) StartInstallAppConfigUpdate(ctx context.Context, input StartInstallAppConfigUpdateInput) (*app.InstallAppConfigVersion, *app.Workflow, error) {
	var install app.Install
	if err := h.db.WithContext(ctx).Where(app.Install{ID: input.InstallID}).First(&install).Error; err != nil {
		return nil, nil, fmt.Errorf("get install: %w", err)
	}
	oldAppConfigID := install.AppConfigID
	if input.AppReleaseID != "" {
		var active app.InstallReleaseDeployment
		err := h.db.WithContext(ctx).Preload("Release").Where(app.InstallReleaseDeployment{
			OrgID: install.OrgID, InstallID: install.ID, Status: app.InstallDeploymentStatusSucceeded,
		}).Order("finished_at DESC, created_at DESC, id DESC").First(&active).Error
		if err == nil {
			oldAppConfigID = active.Release.AppConfigID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("get active install release: %w", err)
		}
	}

	metadata := maps.Clone(input.Metadata)
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["new_app_config_id"] = input.NewAppConfigID
	if input.AppBranchRunID != "" {
		metadata["app_branch_run_id"] = input.AppBranchRunID
	}
	if input.InstallGroupID != "" {
		metadata["install_group_id"] = input.InstallGroupID
	}
	if input.AppReleaseID != "" {
		metadata["app_release_id"] = input.AppReleaseID
		componentBuildIDs, err := json.Marshal(input.ReleaseComponentBuildIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal release component builds: %w", err)
		}
		metadata["release_component_build_ids"] = string(componentBuildIDs)
		metadata["release_sandbox_build_id"] = input.ReleaseSandboxBuildID
	}

	diff, err := configdiff.ComputeInstallConfigDiff(ctx, h.db, oldAppConfigID, input.NewAppConfigID)
	if err != nil {
		return nil, nil, fmt.Errorf("compute install config diff: %w", err)
	}
	diffJSON, err := json.Marshal(diff)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal install config diff: %w", err)
	}

	update := &app.InstallAppConfigVersion{
		InstallID:      input.InstallID,
		OldAppConfigID: oldAppConfigID,
		NewAppConfigID: input.NewAppConfigID,
		Status:         app.NewCompositeStatus(ctx, app.StatusPending),
		Metadata:       metadata,
	}
	if input.AppReleaseID != "" {
		update.AppReleaseID = generics.ToPtr(input.AppReleaseID)
	}
	if input.OperatingModelID != "" {
		update.OperatingModelID = generics.ToPtr(input.OperatingModelID)
	}
	if input.AppBranchRunID != "" {
		update.AppBranchRunID = generics.ToPtr(input.AppBranchRunID)
	}
	if input.InstallGroupID != "" {
		update.InstallGroupID = generics.ToPtr(input.InstallGroupID)
	}
	if err := h.db.WithContext(ctx).Create(update).Error; err != nil {
		return nil, nil, fmt.Errorf("create install app config version: %w", err)
	}
	blobID := domains.NewBlobID()
	s3Key := fmt.Sprintf("blobs/install_config_diffs/%s", blobID)
	checksum, err := h.blobSvc.UploadStream(ctx, s3Key, strings.NewReader(string(diffJSON)))
	if err != nil {
		h.l.Warn("unable to save install config diff blob", zap.Error(err))
	} else {
		diffMetadata, err := json.Marshal(blobstore.BlobMetadata{
			BlobID:      blobID,
			S3Key:       s3Key,
			Size:        int64(len(diffJSON)),
			ContentType: "application/json",
			Checksum:    checksum,
			CreatedAt:   time.Now().Format(time.RFC3339),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("marshal install config diff metadata: %w", err)
		}
		if err := h.db.WithContext(ctx).Model(update).Update("diff", string(diffMetadata)).Error; err != nil {
			return nil, nil, fmt.Errorf("save install config diff metadata: %w", err)
		}
	}
	metadata["install_config_update_id"] = update.ID

	workflow, err := h.CreateWorkflowWithApprovalOption(ctx, input.InstallID, app.WorkflowTypeAppBranchConfigUpdate, metadata, input.PlanOnly, input.ApprovalOption)
	if err != nil {
		return nil, nil, fmt.Errorf("create workflow: %w", err)
	}
	update.WorkflowID = &workflow.ID
	if err := h.db.WithContext(ctx).Model(update).Update("workflow_id", workflow.ID).Error; err != nil {
		return nil, nil, fmt.Errorf("link install app config version to workflow: %w", err)
	}

	queue, err := h.queueClient.GetQueueByOwner(ctx, input.InstallID, "installs")
	if err != nil {
		return nil, nil, fmt.Errorf("resolve install workflow queue: %w", err)
	}
	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queue.ID,
		Signal:    &executeflow.Signal{WorkflowID: workflow.ID},
		OwnerID:   workflow.ID,
		OwnerType: "install_workflows",
		Callback:  input.Callback,
	}); err != nil {
		return nil, nil, fmt.Errorf("enqueue app config update workflow: %w", err)
	}

	return update, workflow, nil
}
