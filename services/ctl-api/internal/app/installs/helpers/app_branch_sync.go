package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
)

// AppBranchRunForInstall describes the app branch run an install should be on.
type AppBranchRunForInstall struct {
	AppBranchRunID               string
	AppConfigID                  string
	InstallGroupID               string
	InstallGroupName             string
	InstallGroupAssignmentSource app.InstallAppBranchGroupAssignmentSource
	AlreadyCurrent               bool
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

var (
	ErrNoDeployableAppBranchRun = errors.New("app branch has no deployable run")
	ErrNoMatchingInstallGroup   = errors.New("install matches no install group on app branch")
)

// LatestAppBranchRunForInstall returns the latest deployable app branch run for
// the branch, along with the install group the install falls into. It returns a
// zero value only when the install is no longer pinned to the branch.
func (h *Helpers) LatestAppBranchRunForInstall(ctx context.Context, appBranchID, installID string) (*AppBranchRunForInstall, error) {
	var install app.Install
	if err := h.db.WithContext(ctx).Where(app.Install{ID: installID}).First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}
	if !install.AppBranchID.Valid || install.AppBranchID.String != appBranchID {
		return &AppBranchRunForInstall{}, nil
	}

	return h.ResolveAppBranchRunForInstall(ctx, appBranchID, &install)
}

func AppBranchRunResolveActivityError(err error) error {
	var userErr stderr.ErrUser
	if errors.As(err, &userErr) {
		return temporal.NewNonRetryableApplicationError(userErr.Description, "AppBranchRunNotDeployable", err)
	}
	return err
}

func (h *Helpers) ResolveAppBranchRunForInstall(ctx context.Context, appBranchID string, install *app.Install) (*AppBranchRunForInstall, error) {
	run, err := h.LatestDeployableAppBranchRun(ctx, appBranchID)
	if err != nil {
		return nil, err
	}

	// Membership follows the branch's latest config (what the API and dashboard
	// expose). Group rows are minted per config version, so a pin to a group
	// added after the last deployable run would miss if we resolved against
	// that run's config.
	groups, err := h.appsHelpers.LatestConfigInstallGroups(ctx, appBranchID)
	if err != nil {
		return nil, err
	}
	group, source, err := appshelpers.ResolveInstallGroupAssignment(groups, install)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("install %s on app branch %s: %w", install.ID, appBranchID, ErrNoMatchingInstallGroup),
			Description: "The install does not match any install group on the selected app branch.",
		}
	}

	installGroupID := ""
	var runGroup app.AppBranchInstallGroup
	err = h.db.WithContext(ctx).
		Where(app.AppBranchInstallGroup{AppBranchConfigID: run.AppBranchConfigID, Name: group.Name}).
		First(&runGroup).Error
	if err == nil {
		installGroupID = runGroup.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to get install group for app branch run: %w", err)
	}

	alreadyCurrent, err := h.installOnAppConfig(ctx, install, run.AppConfigID)
	if err != nil {
		return nil, err
	}

	return &AppBranchRunForInstall{
		AppBranchRunID:               run.ID,
		AppConfigID:                  run.AppConfigID,
		InstallGroupID:               installGroupID,
		InstallGroupName:             group.Name,
		InstallGroupAssignmentSource: source,
		AlreadyCurrent:               alreadyCurrent,
	}, nil
}

func (h *Helpers) LatestDeployableAppBranchRun(ctx context.Context, appBranchID string) (*app.AppBranchRun, error) {
	var run app.AppBranchRun
	err := h.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Joins("JOIN app_configs ON app_configs.id = app_branch_runs.app_config_id AND app_configs.deleted_at = 0").
		Where("app_branch_runs.app_branch_id = ?", appBranchID).
		Where("app_branch_runs.run_type IN ?", []app.AppBranchRunType{app.AppBranchRunTypeGit, app.AppBranchRunTypeManual}).
		Where("app_branch_runs.plan_only = ?", false).
		Where("app_branch_runs.app_config_id != ''").
		Where("app_branch_runs.labels->>? = ?", app.AppBranchRunLabelBuildsCompleted, "true").
		Where("app_configs.status_v2->>'status' = ?", string(app.AppConfigStatusActive)).
		Where("NOT EXISTS (SELECT 1 FROM app_branch_run_previews p WHERE p.app_branch_run_id = app_branch_runs.id AND p.deleted_at = 0)").
		Order("app_branch_runs.created_at DESC").
		First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("app branch %s: %w", appBranchID, ErrNoDeployableAppBranchRun),
			Description: "The selected app branch has no completed, non-preview run to deploy. Trigger a branch run first.",
		}
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get latest app branch run: %w", err)
	}
	return &run, nil
}

func (h *Helpers) installOnAppConfig(ctx context.Context, install *app.Install, appConfigID string) (bool, error) {
	if install.DeployedAppConfigID() == appConfigID {
		return true, nil
	}

	var inFlight int64
	if err := h.db.WithContext(ctx).
		Model(&app.InstallAppConfigVersion{}).
		Joins("JOIN install_workflows ON install_workflows.id = install_app_config_versions.workflow_id").
		Where("install_app_config_versions.install_id = ?", install.ID).
		Where("install_app_config_versions.new_app_config_id = ?", appConfigID).
		Where("install_workflows.plan_only = ?", false).
		Where("install_workflows.finished_at IS NULL").
		Where("install_workflows.status->>'status' NOT IN ?", []string{string(app.StatusError), string(app.StatusCancelled)}).
		Count(&inFlight).Error; err != nil {
		return false, fmt.Errorf("unable to check in-flight app config rollouts: %w", err)
	}
	return inFlight > 0, nil
}

// CreateAppBranchConfigUpdateWorkflow records an install app config version for
// the new config and enqueues the workflow that rolls the install onto it.
func (h *Helpers) CreateAppBranchConfigUpdateWorkflow(ctx context.Context, input AppBranchConfigUpdateInput) (*AppBranchConfigUpdate, error) {
	var install app.Install
	if err := h.db.WithContext(ctx).First(&install, "id = ?", input.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	deployedAppConfigID := install.DeployedAppConfigID()

	diff, err := configdiff.ComputeInstallConfigDiff(ctx, h.db, deployedAppConfigID, input.NewAppConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to compute config diff: %w", err)
	}

	update := app.InstallAppConfigVersion{
		InstallID:      input.InstallID,
		OldAppConfigID: deployedAppConfigID,
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
