package helpers

import (
	"context"
	"fmt"
	"slices"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

// LookupVCSConnection: lookup a VCS connection from a list of vcs conns, for the one that has access to the repo
// provided. Connections are iterated most-recent first so that newly created connections take precedence over
// older ones that may still grant access to the repo (e.g. when a connection has been re-created).
func (h *Helpers) LookupVCSConnection(ctx context.Context,
	owner, name string,
	vcsConnections []app.VCSConnection) (string, error) {
	if len(vcsConnections) < 1 {
		return "", stderr.ErrUser{
			Err:         fmt.Errorf("no vcs connections on org: %w", gorm.ErrRecordNotFound),
			Description: "please create a vcs connection before proceeding",
		}
	}

	// Sort most recent first so we prefer the newest connection with access.
	sortedConns := slices.Clone(vcsConnections)
	slices.SortStableFunc(sortedConns, func(a, b app.VCSConnection) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	})

	for _, vcsConn := range sortedConns {
		client, err := h.GetVCSConnectionClient(ctx, &vcsConn)
		if err != nil {
			return "", fmt.Errorf("unable to get client: %w", err)
		}

		repo, _, err := client.Repositories.Get(ctx, owner, name)
		if err != nil {
			continue
		}

		if *repo.Visibility == "public" {
			return "", stderr.ErrUser{
				Err:         fmt.Errorf("can not use a public repo with a connected_repo config"),
				Description: "please use a `public_repo` block instead",
			}
		}
		return vcsConn.ID, nil
	}

	return "", stderr.ErrUser{
		Err:         fmt.Errorf("no vcs connection found with access to %s/%s", owner, name),
		Description: "please make sure vcs connection has access to this repo",
	}
}
