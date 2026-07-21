package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type GetBuildGitSource struct {
	BuildID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field BuildID
func (a *Activities) GetBuildGitSource(ctx context.Context, req GetBuildGitSource) (*plantypes.GitSource, error) {
	build, err := a.getComponentBuildWithConfig(ctx, req.BuildID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get build config")
	}

	switch build.ComponentConfigConnection.VCSConnectionType {
	case app.VCSConnectionTypeConnectedRepo:
		gitRef := build.GitRef
		if gitRef == nil {
			cfg := build.ComponentConfigConnection.ConnectedGithubVCSConfig
			commit, err := a.vcsHelpers.GetConnectedGithubVCSConfigLatestCommit(ctx, cfg)
			if err != nil {
				return nil, err
			}

			vcsCommit := a.vcsHelpers.GithubCommitToVCSConnectionCommit(
				commit,
				cfg.ID,
				plugins.TableName(a.db, &app.ConnectedGithubVCSConfig{}),
				cfg.VCSConnectionID,
			)
			if vcsCommit == nil {
				return nil, fmt.Errorf("invalid commit data from GitHub")
			}

			resolvedRef := vcsCommit.SHA
			if err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				var current app.ComponentBuild
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where(app.ComponentBuild{ID: build.ID}).
					First(&current).Error; err != nil {
					return fmt.Errorf("unable to lock component build: %w", err)
				}
				if current.GitRef != nil {
					resolvedRef = *current.GitRef
					return nil
				}

				if err := tx.Create(vcsCommit).Error; err != nil {
					return fmt.Errorf("unable to create vcs commit: %w", err)
				}
				if err := tx.Model(&app.ComponentBuild{}).
					Where(app.ComponentBuild{ID: build.ID}).
					Updates(map[string]any{
						"git_ref":                  vcsCommit.SHA,
						"vcs_connection_commit_id": vcsCommit.ID,
					}).Error; err != nil {
					return fmt.Errorf("unable to update component build git source: %w", err)
				}
				return nil
			}); err != nil {
				return nil, err
			}
			gitRef = &resolvedRef
		}

		return a.vcsHelpers.GetGitSourceAtCommit(ctx, build.ComponentConfigConnection.ConnectedGithubVCSConfig, *gitRef)
	case app.VCSConnectionTypePublicRepo:
		return a.vcsHelpers.GetPubliGitSource(ctx, build.ComponentConfigConnection.PublicGitVCSConfig)
	default:
	}

	return nil, nil
}
