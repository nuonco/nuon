package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field runID
func (a *Activities) updateAppBranchRunVCSCommit(ctx context.Context, runID, vcsCommitID string) error {
	var run app.AppBranchRun
	if err := a.db.WithContext(ctx).Where(app.AppBranchRun{ID: runID}).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("app branch run not found: %s: %w", runID, err)
		}
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	var commit app.VCSConnectionCommit
	if err := a.db.WithContext(ctx).Where(app.VCSConnectionCommit{ID: vcsCommitID}).First(&commit).Error; err != nil {
		return fmt.Errorf("unable to get vcs commit: %w", err)
	}

	updated := app.AppBranchRun{VCSConnectionCommitID: &vcsCommitID}
	columns := []string{"vcs_connection_commit_id"}
	if commit.SHA != "" && commit.SHA != run.HeadSHA && persistResolvedPinSHA(&run) {
		meta := run.Metadata
		meta.HeadSHA = commit.SHA
		updated.HeadSHA = commit.SHA
		updated.Metadata = meta
		columns = append(columns, "head_sha", "metadata")
	}

	res := a.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Where(app.AppBranchRun{ID: runID}).
		Select(columns).
		Updates(&updated)
	if res.Error != nil {
		return fmt.Errorf("unable to update run VCS commit: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("app branch run not found: %s: %w", runID, gorm.ErrRecordNotFound)
	}
	return nil
}

func persistResolvedPinSHA(run *app.AppBranchRun) bool {
	if run == nil || run.RunType != app.AppBranchRunTypeManual {
		return false
	}
	meta := run.RunMetadata()
	return meta.Tag != "" || meta.GitRef != "" || run.HeadSHA != "" || meta.PRNumber != nil
}
