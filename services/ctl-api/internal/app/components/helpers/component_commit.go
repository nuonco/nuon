package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// GetComponentCommit will return a commit for a component, when a connected git source is attached.
func (s *Helpers) GetComponentCommit(ctx context.Context, cmpID string) (*app.VCSConnectionCommit, error) {
	cmp, err := s.GetComponent(ctx, cmpID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	if cmp.LatestConfig.VCSConnectionType != app.VCSConnectionTypeConnectedRepo {
		return nil, fmt.Errorf("unable to get component config type for non connected-repo vcs configs")
	}

	// find the latest commit for this connection
	commit, err := s.vcsHelpers.GetVCSConfigLatestCommit(ctx, cmp.LatestConfig.ConnectedGithubVCSConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to get the latest commit: %w", err)
	}
	vcsCommit := app.VCSConnectionCommit{
		SHA:             *commit.SHA,
		Message:         *commit.Commit.Message,
		VCSConnectionID: cmp.LatestConfig.ConnectedGithubVCSConfig.VCSConnectionID,
		AuthorName:      generics.FromPtrStr(commit.Author.Name),
		AuthorEmail:     generics.FromPtrStr(commit.Author.Email),
	}

	res := s.db.WithContext(ctx).Create(&vcsCommit)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create vcs commit: %w", res.Error)
	}

	return &vcsCommit, nil
}
