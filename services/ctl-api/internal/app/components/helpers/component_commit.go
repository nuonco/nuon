package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

func (s *Helpers) GetComponentCommit(ctx context.Context, cmpID string) (*app.VCSConnectionCommit, error) {
	cmp, err := s.GetComponent(ctx, cmpID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	if cmp.LatestConfig.VCSConnectionType != app.VCSConnectionTypeConnectedRepo {
		return nil, fmt.Errorf("unable to get component config type for non connected-repo vcs configs")
	}

	commit, err := s.vcsHelpers.GetConnectedGithubVCSConfigLatestCommit(ctx, cmp.LatestConfig.ConnectedGithubVCSConfig)
	if err != nil {
		return nil, err
	}

	vcsCommit := s.vcsHelpers.GithubCommitToVCSConnectionCommit(commit,
		cmp.LatestConfig.ConnectedGithubVCSConfig.ID,
		plugins.TableName(s.db, &app.ConnectedGithubVCSConfig{}),
		cmp.LatestConfig.ConnectedGithubVCSConfig.VCSConnectionID)
	if vcsCommit == nil {
		return nil, fmt.Errorf("invalid commit data from GitHub")
	}

	res := s.db.WithContext(ctx).Create(vcsCommit)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create vcs commit: %w", res.Error)
	}

	return vcsCommit, nil
}
