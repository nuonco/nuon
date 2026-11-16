package helpers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"gorm.io/gorm"

	"github.com/google/go-github/v50/github"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// LookupVCSConnection returns the org VCS connection for the GitHub account that
// owns the repository. Public repositories must still use that owner connection
// rather than the first installation whose token can read the repo.
func (h *Helpers) LookupVCSConnection(ctx context.Context,
	owner, name string,
	vcsConnections []app.VCSConnection) (string, error) {
	if len(vcsConnections) < 1 {
		return "", stderr.ErrUser{
			Err:         fmt.Errorf("no vcs connections on org: %w", gorm.ErrRecordNotFound),
			Description: "please create a vcs connection before proceeding",
		}
	}

	conn := connectionForRepoOwner(owner, vcsConnections)
	if conn == nil {
		return "", stderr.ErrUser{
			Err:         fmt.Errorf("no vcs connection for github account %s", owner),
			Description: fmt.Sprintf("connect the %s GitHub account before using %s/%s", owner, owner, name),
		}
	}

	if err := h.repoAccess(ctx, conn, owner, name); err != nil {
		return "", fmt.Errorf("vcs connection %s cannot access %s/%s: %w", conn.ID, owner, name, err)
	}

	return conn.ID, nil
}

type repoAccessCacheKey struct{}

func WithRepoAccessCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, repoAccessCacheKey{}, &sync.Map{})
}

func connectionForRepoOwner(owner string, vcsConnections []app.VCSConnection) *app.VCSConnection {
	for i := range vcsConnections {
		if strings.EqualFold(vcsConnections[i].GithubAccountName, owner) {
			return &vcsConnections[i]
		}
	}
	return nil
}

func (h *Helpers) repoAccess(ctx context.Context, conn *app.VCSConnection, owner, name string) error {
	cache, _ := ctx.Value(repoAccessCacheKey{}).(*sync.Map)
	key := strings.Join([]string{conn.OrgID, conn.ID, conn.GithubInstallID, strings.ToLower(owner), strings.ToLower(name)}, "\n")
	if cache != nil {
		if _, ok := cache.Load(key); ok {
			return nil
		}
	}

	client, err := h.GetVCSConnectionClient(ctx, conn)
	if err != nil {
		var notFound stderr.ErrNotFound
		if errors.As(err, &notFound) {
			return stderr.ErrUser{Err: err, Description: notFound.Description}
		}
		return fmt.Errorf("unable to get client: %w", err)
	}

	if _, _, err := client.Repositories.Get(ctx, owner, name); err != nil {
		var ghErr *github.ErrorResponse
		if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
			return stderr.ErrUser{
				Err:         err,
				Description: fmt.Sprintf("please make sure the %s GitHub connection has access to %s/%s", owner, owner, name),
			}
		}
		return err
	}
	if cache != nil {
		cache.Store(key, struct{}{})
	}
	return nil
}
