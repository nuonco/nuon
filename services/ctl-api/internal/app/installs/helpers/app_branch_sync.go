package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
)

// AppBranchRunForInstall describes the app branch run an install should be on.
type AppBranchRunForInstall struct {
	AppBranchRunID string
	AppConfigID    string
	InstallGroupID string
	AlreadyCurrent bool
}

// AppBranchConfigUpdateInput describes the config update to run for an install.
type AppBranchConfigUpdateInput struct {
	InstallID      string
	NewAppConfigID string
	AppBranchRunID string
	InstallGroupID string
	PlanOnly       bool
	Callback       callback.Ref
}

// AppBranchConfigUpdate is the workflow and version created for a config update.
type AppBranchConfigUpdate struct {
	WorkflowID                string
	InstallAppConfigVersionID string
}

// EnsureInstallAppBranch pins the install to the app branch, if it isn't already.
func (h *Helpers) EnsureInstallAppBranch(ctx context.Context, installID, appBranchID string) error {
	var install app.Install
	if err := h.db.WithContext(ctx).Where(app.Install{ID: installID}).First(&install).Error; err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	var branch app.AppBranch
	if err := h.db.WithContext(ctx).Where(app.AppBranch{ID: appBranchID, AppID: install.AppID}).First(&branch).Error; err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	if install.AppBranchID.Valid && install.AppBranchID.String == branch.ID {
		return nil
	}
	if err := h.appsHelpers.SetInstallAppBranch(ctx, install.ID, branch.ID); err != nil {
		return fmt.Errorf("unable to move install to app branch: %w", err)
	}
	return nil
}

// LatestAppBranchRunForInstall returns the latest non-preview app branch run for
// the branch, along with the install group the install falls into. It returns a
// zero value when the install is no longer pinned to the branch or no run has
// produced an app config yet.
func (h *Helpers) LatestAppBranchRunForInstall(ctx context.Context, appBranchID, installID string) (*AppBranchRunForInstall, error) {
	var install app.Install
	if err := h.db.WithContext(ctx).Where(app.Install{ID: installID}).First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}
	if !install.AppBranchID.Valid || install.AppBranchID.String != appBranchID {
		return &AppBranchRunForInstall{}, nil
	}

	var run app.AppBranchRun
	err := h.db.WithContext(ctx).
		Where(app.AppBranchRun{AppBranchID: appBranchID}).
		Where("run_type != ?", app.AppBranchRunTypeGitPreview).
		Where("app_config_id != ''").
		Order("created_at DESC").
		First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &AppBranchRunForInstall{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get latest app branch run: %w", err)
	}

	var groups []app.AppBranchInstallGroup
	if err := h.db.WithContext(ctx).
		Where(app.AppBranchInstallGroup{AppBranchConfigID: run.AppBranchConfigID}).
		Find(&groups).Error; err != nil {
		return nil, fmt.Errorf("unable to get install groups for app branch run: %w", err)
	}

	var installGroupID string
	for i := range groups {
		if appshelpers.InstallMatchesGroup(&groups[i], &install) {
			installGroupID = groups[i].ID
			break
		}
	}

	return &AppBranchRunForInstall{
		AppBranchRunID: run.ID,
		AppConfigID:    run.AppConfigID,
		InstallGroupID: installGroupID,
		AlreadyCurrent: install.AppConfigID == run.AppConfigID,
	}, nil
}

// CreateAppBranchConfigUpdateWorkflow records an install app config version for
// the new config and enqueues the workflow that rolls the install onto it.
func (h *Helpers) CreateAppBranchConfigUpdateWorkflow(ctx context.Context, input AppBranchConfigUpdateInput) (*AppBranchConfigUpdate, error) {
	var install app.Install
	if err := h.db.WithContext(ctx).First(&install, "id = ?", input.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	diff, err := configdiff.ComputeInstallConfigDiff(ctx, h.db, install.AppConfigID, input.NewAppConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to compute config diff: %w", err)
	}

	update := app.InstallAppConfigVersion{
		InstallID:      input.InstallID,
		OldAppConfigID: install.AppConfigID,
		NewAppConfigID: input.NewAppConfigID,
		Status:         app.NewCompositeStatus(ctx, app.StatusPending),
	}
	if input.AppBranchRunID != "" {
		update.AppBranchRunID = &input.AppBranchRunID
	}
	if input.InstallGroupID != "" {
		update.InstallGroupID = &input.InstallGroupID
	}
	if err := h.db.WithContext(ctx).Create(&update).Error; err != nil {
		return nil, fmt.Errorf("unable to create install config update: %w", err)
	}

	if err := h.SaveInstallConfigDiffBlob(ctx, update.ID, diff); err != nil {
		h.l.Warn("unable to save config diff blob", zap.Error(err))
	}

	metadata := map[string]string{
		"new_app_config_id":        input.NewAppConfigID,
		"install_config_update_id": update.ID,
	}
	if input.AppBranchRunID != "" {
		metadata["app_branch_run_id"] = input.AppBranchRunID
	}
	if input.InstallGroupID != "" {
		metadata["install_group_id"] = input.InstallGroupID
	}

	wf, err := h.CreateWorkflow(ctx, input.InstallID, app.WorkflowTypeAppBranchConfigUpdate, metadata, input.PlanOnly)
	if err != nil {
		return nil, fmt.Errorf("unable to create install config update workflow: %w", err)
	}

	if err := h.db.WithContext(ctx).
		Model(&update).
		Update("workflow_id", wf.ID).Error; err != nil {
		return nil, fmt.Errorf("unable to link workflow to install config update: %w", err)
	}

	if err := h.EnqueueInstallWorkflowWithCallback(ctx, input.InstallID, wf.ID, input.Callback); err != nil {
		return nil, fmt.Errorf("unable to enqueue workflow for install %s: %w", input.InstallID, err)
	}

	return &AppBranchConfigUpdate{
		WorkflowID:                wf.ID,
		InstallAppConfigVersionID: update.ID,
	}, nil
}

// SaveInstallConfigDiffBlob uploads the diff and records its blob metadata on
// the install app config version.
func (h *Helpers) SaveInstallConfigDiffBlob(ctx context.Context, installConfigVersionID string, diff *app.InstallConfigDiff) error {
	diffJSON, err := json.Marshal(diff)
	if err != nil {
		return fmt.Errorf("unable to marshal diff: %w", err)
	}

	blobID := domains.NewBlobID()
	s3Key := fmt.Sprintf("blobs/install_config_diffs/%s", blobID)

	checksum, err := h.blobSvc.UploadStream(ctx, s3Key, strings.NewReader(string(diffJSON)))
	if err != nil {
		return fmt.Errorf("unable to upload diff to S3: %w", err)
	}

	metadataJSON, err := json.Marshal(blobstore.BlobMetadata{
		BlobID:      blobID,
		S3Key:       s3Key,
		Size:        int64(len(diffJSON)),
		ContentType: "application/json",
		Checksum:    checksum,
		CreatedAt:   time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("unable to marshal blob metadata: %w", err)
	}

	res := h.db.WithContext(ctx).
		Model(&app.InstallAppConfigVersion{}).
		Where(app.InstallAppConfigVersion{ID: installConfigVersionID}).
		Update("diff", string(metadataJSON))
	if res.Error != nil {
		return fmt.Errorf("unable to save diff: %w", res.Error)
	}

	return nil
}
