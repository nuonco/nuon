package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix AppBranches
// @by-field vcsConfigID
func (a *Activities) createOrGetVCSCommit(ctx context.Context, vcsConfigID string) (*app.VCSConnectionCommit, error) {
	vcsHelpers := a.helpers.VCSHelpers()

	// Try ConnectedGithubVCSConfig first
	var connectedCfg app.ConnectedGithubVCSConfig
	res := a.db.WithContext(ctx).
		Preload("VCSConnection").
		First(&connectedCfg, "id = ?", vcsConfigID)

	if res.Error == nil {
		// Get the latest commit from GitHub
		ghCommit, err := vcsHelpers.GetVCSConfigLatestCommit(ctx, &connectedCfg)
		if err != nil {
			return nil, fmt.Errorf("unable to get latest commit for connected repo: %w", err)
		}

		// Create or find the VCSConnectionCommit record
		vcsCommit := &app.VCSConnectionCommit{
			VCSConnectionID: connectedCfg.VCSConnectionID,
			SHA:             *ghCommit.SHA,
		}

		// Add commit metadata if available
		if ghCommit.Commit != nil {
			if ghCommit.Commit.Author != nil {
				if ghCommit.Commit.Author.Name != nil {
					vcsCommit.AuthorName = *ghCommit.Commit.Author.Name
				}
				if ghCommit.Commit.Author.Email != nil {
					vcsCommit.AuthorEmail = *ghCommit.Commit.Author.Email
				}
			}
			if ghCommit.Commit.Message != nil {
				vcsCommit.Message = *ghCommit.Commit.Message
			}
		}

		// First try to find existing commit by SHA and VCS connection
		var existingCommit app.VCSConnectionCommit
		findRes := a.db.WithContext(ctx).
			Where("sha = ? AND vcs_connection_id = ?", vcsCommit.SHA, vcsCommit.VCSConnectionID).
			First(&existingCommit)

		if findRes.Error == nil {
			// Found existing commit, return it
			return &existingCommit, nil
		}

		// Create new commit record
		createRes := a.db.WithContext(ctx).Create(vcsCommit)
		if createRes.Error != nil {
			return nil, fmt.Errorf("unable to create VCS commit record: %w", createRes.Error)
		}

		return vcsCommit, nil
	}

	// Try PublicGitVCSConfig
	var publicCfg app.PublicGitVCSConfig
	res = a.db.WithContext(ctx).First(&publicCfg, "id = ?", vcsConfigID)
	if res.Error == nil {
		// Get the latest commit from GitHub
		ghCommit, err := vcsHelpers.GetPublicGitVCSConfigLatestCommit(ctx, &publicCfg)
		if err != nil {
			return nil, fmt.Errorf("unable to get latest commit for public repo: %w", err)
		}

		// For public repos, we don't have a VCSConnection, so we need to handle this differently
		// We'll create a commit record without a VCSConnectionID
		vcsCommit := &app.VCSConnectionCommit{
			SHA: *ghCommit.SHA,
		}

		// Add commit metadata if available
		if ghCommit.Commit != nil {
			if ghCommit.Commit.Author != nil {
				if ghCommit.Commit.Author.Name != nil {
					vcsCommit.AuthorName = *ghCommit.Commit.Author.Name
				}
				if ghCommit.Commit.Author.Email != nil {
					vcsCommit.AuthorEmail = *ghCommit.Commit.Author.Email
				}
			}
			if ghCommit.Commit.Message != nil {
				vcsCommit.Message = *ghCommit.Commit.Message
			}
		}

		// For public repos without VCS connection, find by SHA only
		var existingCommit app.VCSConnectionCommit
		findRes := a.db.WithContext(ctx).
			Where("sha = ? AND vcs_connection_id IS NULL", vcsCommit.SHA).
			First(&existingCommit)

		if findRes.Error == nil {
			// Found existing commit, return it
			return &existingCommit, nil
		}

		// Create new commit record
		createRes := a.db.WithContext(ctx).Create(vcsCommit)
		if createRes.Error != nil {
			return nil, fmt.Errorf("unable to create VCS commit record: %w", createRes.Error)
		}

		return vcsCommit, nil
	}

	return nil, fmt.Errorf("VCS config not found: %s", vcsConfigID)
}
