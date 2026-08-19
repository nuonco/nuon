package helpers

import (
	"context"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func (h *Helpers) CreateAppBranchConfig(
	ctx context.Context,
	appBranchID string,
	connectedGithubVCSConfig *app.ConnectedGithubVCSConfig,
	publicGitVCSConfig *app.PublicGitVCSConfig,
	installGroups []app.AppBranchInstallGroup,
	postDeployRunbookIDs *[]string,
) (*app.AppBranchConfig, error) {
	return h.CreateAppBranchConfigWithDB(ctx, h.db, appBranchID, connectedGithubVCSConfig, publicGitVCSConfig, installGroups, postDeployRunbookIDs)
}

// Callers inside a transaction must use this, or the app_branch_id FK fails.
func (h *Helpers) CreateAppBranchConfigWithDB(
	ctx context.Context,
	db *gorm.DB,
	appBranchID string,
	connectedGithubVCSConfig *app.ConnectedGithubVCSConfig,
	publicGitVCSConfig *app.PublicGitVCSConfig,
	installGroups []app.AppBranchInstallGroup,
	postDeployRunbookIDs *[]string,
) (*app.AppBranchConfig, error) {
	if postDeployRunbookIDs != nil && len(*postDeployRunbookIDs) > 0 {
		if err := h.validatePostDeployRunbooks(ctx, db, appBranchID, *postDeployRunbookIDs); err != nil {
			return nil, err
		}
	}

	config := app.AppBranchConfig{
		AppBranchID:              appBranchID,
		InstallGroups:            installGroups,
		ConnectedGithubVCSConfig: connectedGithubVCSConfig,
		PublicGitVCSConfig:       publicGitVCSConfig,
	}

	var previous app.AppBranchConfig
	res := db.WithContext(ctx).
		Where(app.AppBranchConfig{AppBranchID: appBranchID}).
		Order("created_at DESC").
		First(&previous)
	hasPrevious := res.Error == nil
	if hasPrevious {
		config.ComponentIDs = previous.ComponentIDs
		config.ActionIDs = previous.ActionIDs
		config.RunbookIDs = previous.RunbookIDs
	} else if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to load previous app branch config: %w", res.Error)
	}

	// Each call mints a new config row, and several callers (deployment plan
	// editor, branch rename, install branch connection) legitimately know nothing
	// about post-deploy runbooks. A nil slice therefore means "carry forward"; an
	// explicitly empty one means "clear".
	if postDeployRunbookIDs != nil {
		config.PostDeployRunbookIDs = pq.StringArray(*postDeployRunbookIDs)
	} else if hasPrevious {
		config.PostDeployRunbookIDs = previous.PostDeployRunbookIDs
	}

	if err := db.WithContext(ctx).Create(&config).Error; err != nil {
		return nil, fmt.Errorf("unable to create app branch config: %w", err)
	}

	return &config, nil
}

// validatePostDeployRunbooks rejects runbook IDs that don't belong to the branch's
// app. Every caller passes through here, so a bad ID fails at config time rather
// than mid-rollout.
func (h *Helpers) validatePostDeployRunbooks(ctx context.Context, db *gorm.DB, appBranchID string, runbookIDs []string) error {
	var branch app.AppBranch
	if err := db.WithContext(ctx).First(&branch, "id = ?", appBranchID).Error; err != nil {
		return fmt.Errorf("unable to find app branch: %w", err)
	}

	var found []app.Runbook
	if err := db.WithContext(ctx).
		Where(app.Runbook{AppID: branch.AppID}).
		Where("id IN ?", runbookIDs).
		Find(&found).Error; err != nil {
		return fmt.Errorf("unable to load post-deploy runbooks: %w", err)
	}

	foundIDs := make(map[string]struct{}, len(found))
	for _, rb := range found {
		foundIDs[rb.ID] = struct{}{}
	}

	for _, id := range runbookIDs {
		if _, ok := foundIDs[id]; !ok {
			return stderr.ErrUser{
				Err:         fmt.Errorf("runbook %q does not belong to app %q", id, branch.AppID),
				Description: fmt.Sprintf("post-deploy runbook %q does not belong to this app", id),
			}
		}
	}

	return nil
}
