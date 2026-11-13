package account

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type DeleteDeadTokensRequest struct {
	AfterID       string
	Limit         int
	RevokedBefore time.Time
	ExpiredBefore time.Time
	DryRun        bool
}

type DeleteDeadTokensResult struct {
	Selected int    `json:"selected"`
	Deleted  int    `json:"deleted"`
	LastID   string `json:"last_id"`
}

// DeleteDeadTokens hard-deletes tokens revoked before RevokedBefore and unrevoked
// tokens that expired before ExpiredBefore. Expired static tokens are kept because
// the token list endpoints still show them. A token that a live OAuth grant names
// as its source is kept too: redeeming the grant requires the source row to exist,
// even when it is revoked or expired.
func (c *Client) DeleteDeadTokens(ctx context.Context, req DeleteDeadTokensRequest) (*DeleteDeadTokensResult, error) {
	out := &DeleteDeadTokensResult{}
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET LOCAL plan_cache_mode = force_custom_plan").Error; err != nil {
			return errors.Wrap(err, "unable to force custom query plans")
		}

		liveGrants := func(model any) *gorm.DB {
			return tx.Model(model).Select("1").Where("source_token_id = tokens.id")
		}
		query := tx.Model(&app.Token{}).
			Unscoped().
			Where("tokens.id > ?", req.AfterID).
			Where(tx.Where("tokens.deleted_at BETWEEN 1 AND ?", req.RevokedBefore.Unix()).
				Or("tokens.deleted_at = 0 AND tokens.expires_at < ? AND tokens.token_type <> ?", req.ExpiredBefore, app.TokenTypeStatic)).
			Where("NOT EXISTS (?)", liveGrants(&app.OAuthRefreshToken{})).
			Where("NOT EXISTS (?)", liveGrants(&app.OAuthAuthorizationCode{})).
			Where("NOT EXISTS (?)", liveGrants(&app.DeviceAuthorization{})).
			Order("tokens.id").
			Limit(req.Limit)
		if !req.DryRun {
			query = query.Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "tokens"}, Options: "SKIP LOCKED"})
		}

		var ids []string
		if err := query.Pluck("tokens.id", &ids).Error; err != nil {
			return errors.Wrap(err, "unable to select dead tokens")
		}
		out.Selected = len(ids)
		if len(ids) == 0 {
			return nil
		}
		out.LastID = ids[len(ids)-1]
		if req.DryRun {
			return nil
		}

		res := tx.Unscoped().Where("id IN ?", ids).Delete(&app.Token{})
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to delete dead tokens")
		}
		out.Deleted = int(res.RowsAffected)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
