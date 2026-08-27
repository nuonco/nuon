package helpers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v50/github"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ResolvePublicRepoGithubClient returns a GitHub client for a public repo.
// When the org has a VCSConnection matching the repo owner, it uses an
// installation token. Otherwise it falls back to an unauthenticated client
// and logs a warning.
func (h *Helpers) ResolvePublicRepoGithubClient(ctx context.Context, l *zap.Logger, orgID, repoOwner string) (*github.Client, bool, error) {
	conn, found, err := h.findOrgVCSConnectionForOwner(ctx, orgID, repoOwner)
	if err != nil {
		return nil, false, err
	}
	if !found {
		h.warnMissingPublicVCSConnection(l, orgID, repoOwner)
		return github.NewClient(nil), false, nil
	}

	client, err := h.GetVCSConnectionClient(ctx, conn)
	if err != nil {
		if l != nil {
			l.Warn("unable to get VCS client for public repo; falling back to unauthenticated client",
				zap.String("org_id", orgID),
				zap.String("repo_owner", repoOwner),
				zap.Error(err),
			)
		}
		return github.NewClient(nil), false, nil
	}
	return client, true, nil
}

// ResolvePublicRepoCloneToken returns an installation token for cloning a public
// repo when the org has a matching VCSConnection. Returns ("", false, nil) when
// falling back to unauthenticated clone (with a warning).
func (h *Helpers) ResolvePublicRepoCloneToken(ctx context.Context, l *zap.Logger, orgID, repoOwner, repoName string) (token string, authenticated bool, err error) {
	conn, found, err := h.findOrgVCSConnectionForOwner(ctx, orgID, repoOwner)
	if err != nil {
		return "", false, err
	}
	if !found {
		h.warnMissingPublicVCSConnection(l, orgID, repoOwner)
		return "", false, nil
	}

	token, err = h.CreateInstallationToken(ctx, conn, repoName)
	if err != nil {
		// Installation may not list this public repo; try a full install token.
		fullToken, fullErr := h.createFullInstallationToken(ctx, conn)
		if fullErr != nil {
			if l != nil {
				l.Warn("unable to create installation token for public repo; falling back to unauthenticated clone",
					zap.String("org_id", orgID),
					zap.String("repo_owner", repoOwner),
					zap.String("repo_name", repoName),
					zap.Error(err),
				)
			}
			return "", false, nil
		}
		return fullToken, true, nil
	}
	return token, true, nil
}

func (h *Helpers) findOrgVCSConnectionForOwner(ctx context.Context, orgID, repoOwner string) (*app.VCSConnection, bool, error) {
	var vcsConn app.VCSConnection
	err := h.db.WithContext(ctx).
		Where(app.VCSConnection{
			OrgID:             orgID,
			GithubAccountName: repoOwner,
		}).
		First(&vcsConn).Error
	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("unable to look up VCS connection for %s: %w", repoOwner, err)
	}
	return &vcsConn, true, nil
}

func (h *Helpers) createFullInstallationToken(ctx context.Context, vcsConn *app.VCSConnection) (string, error) {
	installID, err := strconv.ParseInt(vcsConn.GithubInstallID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("unable to parse github install id: %w", err)
	}
	resp, _, err := h.ghClient.Apps.CreateInstallationToken(ctx, installID, &github.InstallationTokenOptions{})
	if err != nil {
		return "", fmt.Errorf("error creating installation token: %w", err)
	}
	return *resp.Token, nil
}

func (h *Helpers) warnMissingPublicVCSConnection(l *zap.Logger, orgID, repoOwner string) {
	if l == nil {
		return
	}
	l.Warn("no VCS connection for public repo owner; using unauthenticated GitHub access",
		zap.String("org_id", orgID),
		zap.String("repo_owner", repoOwner),
	)
}
