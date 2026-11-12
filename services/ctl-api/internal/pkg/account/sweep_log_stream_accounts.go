package account

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

type SweepLogStreamAccountsRequest struct {
	AfterID          string
	Limit            int
	IdleBefore       time.Time
	OpenStreamBefore time.Time
	DryRun           bool
}

type SweepLogStreamAccountsResult struct {
	Selected int    `json:"selected"`
	Deleted  int    `json:"deleted"`
	Skipped  int    `json:"skipped"`
	LastID   string `json:"last_id"`
}

func (c *Client) SweepLogStreamAccounts(ctx context.Context, req SweepLogStreamAccountsRequest) (*SweepLogStreamAccountsResult, error) {
	out := &SweepLogStreamAccountsResult{}
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Table("accounts").
			Joins("LEFT JOIN log_streams ON log_streams.id = accounts.subject AND log_streams.deleted_at = 0").
			Where("accounts.deleted_at = 0 AND accounts.account_type = 'service' AND accounts.email LIKE 'log%@serviceaccount.nuon.co'").
			Where("accounts.id > ?", req.AfterID).
			Where("accounts.updated_at < ?", req.IdleBefore).
			Where("NOT EXISTS (SELECT 1 FROM tokens WHERE tokens.account_id = accounts.id AND tokens.deleted_at = 0 AND tokens.expires_at > ?)", time.Now()).
			Where("NOT EXISTS (SELECT 1 FROM managed_service_accounts WHERE managed_service_accounts.account_id = accounts.id)").
			Where("log_streams.id IS NULL OR NOT log_streams.open OR log_streams.created_at < ?", req.OpenStreamBefore).
			Order("accounts.id").
			Limit(req.Limit)
		if !req.DryRun {
			query = query.Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "accounts"}, Options: "SKIP LOCKED"})
		}
		var candidates []string
		if err := query.Pluck("accounts.id", &candidates).Error; err != nil {
			return errors.Wrap(err, "unable to select idle log stream accounts")
		}
		out.Selected = len(candidates)
		if len(candidates) == 0 {
			return nil
		}
		out.LastID = candidates[len(candidates)-1]

		var issued []string
		if err := tx.Model(&app.Token{}).
			Where("account_id IN ? AND expires_at > ?", candidates, time.Now()).
			Distinct("account_id").
			Pluck("account_id", &issued).Error; err != nil {
			return errors.Wrap(err, "unable to recheck log stream account tokens")
		}
		skip := make(map[string]struct{}, len(issued))
		for _, id := range issued {
			skip[id] = struct{}{}
		}
		ids := make([]string, 0, len(candidates))
		for _, id := range candidates {
			if _, ok := skip[id]; !ok {
				ids = append(ids, id)
			}
		}
		blocked, err := authz.PrivateRoleBlockedAccounts(tx, ids)
		if err != nil {
			return err
		}
		for id := range blocked {
			skip[id] = struct{}{}
		}

		ids = ids[:0]
		for _, id := range candidates {
			if _, ok := skip[id]; !ok {
				ids = append(ids, id)
			}
		}
		out.Skipped = len(skip)
		if req.DryRun || len(ids) == 0 {
			return nil
		}

		if err := deleteServiceAccounts(tx, ids); err != nil {
			return err
		}
		out.Deleted = len(ids)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
