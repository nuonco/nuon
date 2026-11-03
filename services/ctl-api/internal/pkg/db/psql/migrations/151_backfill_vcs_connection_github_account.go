package migrations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-github/v50/github"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type githubAccount struct {
	id   string
	name string
}

type githubAccountLookup struct {
	acct githubAccount
	err  error
}

func (m *Migrations) Migration151BackfillVCSConnectionGithubAccount(ctx context.Context, db *gorm.DB) error {
	const batchSize = 100

	cache := map[string]githubAccountLookup{}
	var updated, skipped int
	lastID := ""
	for {
		var conns []app.VCSConnection
		if err := db.WithContext(ctx).
			Select(
				"vcs_connections.id",
				"vcs_connections.org_id",
				"vcs_connections.github_install_id",
				"vcs_connections.github_account_id",
				"vcs_connections.github_account_name",
			).
			Joins("JOIN orgs ON orgs.id = vcs_connections.org_id AND orgs.deleted_at = 0").
			Where("vcs_connections.github_account_name = ? OR vcs_connections.github_account_id = ?", "", "").
			Where("vcs_connections.id > ?", lastID).
			Order("vcs_connections.id").
			Limit(batchSize).
			Find(&conns).Error; err != nil {
			return fmt.Errorf("load vcs connections: %w", err)
		}
		if len(conns) == 0 {
			break
		}

		for i := range conns {
			conn := &conns[i]
			acct, skip, err := m.lookupGithubAccount(ctx, cache, conn)
			if err != nil {
				return err
			}
			if skip {
				skipped++
				continue
			}
			wrote, err := writeVCSGithubAccount(ctx, db, conn, acct)
			if err != nil {
				return err
			}
			if wrote {
				updated++
			}
		}
		lastID = conns[len(conns)-1].ID
	}

	if m.l != nil {
		m.l.Info("backfilled vcs connection github accounts",
			zap.Int("updated", updated),
			zap.Int("skipped", skipped),
		)
	}
	return nil
}

func (m *Migrations) lookupGithubAccount(
	ctx context.Context,
	cache map[string]githubAccountLookup,
	conn *app.VCSConnection,
) (githubAccount, bool, error) {
	installID := conn.GithubInstallID
	if cached, ok := cache[installID]; ok {
		if cached.err != nil {
			m.warnSkippedVCSAccount(conn, cached.err)
			return githubAccount{}, true, nil
		}
		return cached.acct, false, nil
	}

	if m.vcsHelpers == nil {
		return githubAccount{}, false, fmt.Errorf("vcs helpers are not configured")
	}

	user, err := m.vcsHelpers.GetInstallationAccount(ctx, installID)
	if err != nil {
		if !skipGithubAccountLookup(err) {
			return githubAccount{}, false, fmt.Errorf("github installation %s: %w", installID, err)
		}
		cache[installID] = githubAccountLookup{err: err}
		m.warnSkippedVCSAccount(conn, err)
		return githubAccount{}, true, nil
	}

	name := user.GetLogin()
	id := user.GetID()
	if name == "" || id == 0 {
		err := fmt.Errorf("github installation %s returned no account login or id", installID)
		cache[installID] = githubAccountLookup{err: err}
		m.warnSkippedVCSAccount(conn, err)
		return githubAccount{}, true, nil
	}

	acct := githubAccount{id: strconv.FormatInt(id, 10), name: name}
	cache[installID] = githubAccountLookup{acct: acct}
	return acct, false, nil
}

func (m *Migrations) warnSkippedVCSAccount(conn *app.VCSConnection, err error) {
	if m.l == nil {
		return
	}
	m.l.Warn("skipping vcs connection github account backfill",
		zap.String("vcs_connection_id", conn.ID),
		zap.String("org_id", conn.OrgID),
		zap.String("github_install_id", conn.GithubInstallID),
		zap.Error(err),
	)
}

func writeVCSGithubAccount(ctx context.Context, db *gorm.DB, conn *app.VCSConnection, acct githubAccount) (bool, error) {
	wrote := false
	if conn.GithubAccountName == "" {
		res := db.WithContext(ctx).
			Model(&app.VCSConnection{}).
			Where("id = ? AND github_account_name = ?", conn.ID, "").
			Update("github_account_name", acct.name)
		if res.Error != nil {
			return false, fmt.Errorf("update github_account_name for %s: %w", conn.ID, res.Error)
		}
		wrote = wrote || res.RowsAffected > 0
	}
	if conn.GithubAccountID == "" {
		res := db.WithContext(ctx).
			Model(&app.VCSConnection{}).
			Where("id = ? AND github_account_id = ?", conn.ID, "").
			Update("github_account_id", acct.id)
		if res.Error != nil {
			return false, fmt.Errorf("update github_account_id for %s: %w", conn.ID, res.Error)
		}
		wrote = wrote || res.RowsAffected > 0
	}
	return wrote, nil
}

func skipGithubAccountLookup(err error) bool {
	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
		return true
	}
	var numErr *strconv.NumError
	if errors.As(err, &numErr) {
		return true
	}
	return strings.Contains(err.Error(), "github installation account is nil")
}
