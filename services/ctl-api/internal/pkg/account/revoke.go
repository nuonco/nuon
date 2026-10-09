package account

import (
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func revokeAccountCredentials(tx *gorm.DB, accountID string) error {
	var acct app.Account
	if err := tx.Unscoped().Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id").
		Where(app.Account{ID: accountID}).
		First(&acct).Error; err != nil {
		return errors.Wrap(err, "unable to lock account for revocation")
	}

	clients := tx.Model(&app.OAuthClient{}).
		Select("id").Where(app.OAuthClient{AccountID: generics.NewNullString(accountID)})
	if res := tx.Where("client_id IN (?)", clients).Delete(&app.OAuthClientSecret{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to revoke account oauth client secrets")
	}

	if res := tx.Where(app.Token{AccountID: accountID}).Delete(&app.Token{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to revoke account tokens")
	}
	if res := tx.Where(app.OAuthRefreshToken{AccountID: accountID}).Delete(&app.OAuthRefreshToken{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to revoke account refresh tokens")
	}
	if res := tx.Where(app.OAuthAuthorizationCode{AccountID: accountID}).Delete(&app.OAuthAuthorizationCode{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to revoke account authorization codes")
	}
	if res := tx.Where(app.DeviceCode{AccountID: accountID}).Delete(&app.DeviceCode{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to revoke account device codes")
	}
	return nil
}

func RevokeDerivedCredentials(tx *gorm.DB, sourceType app.TokenSourceType, sourceIDs []string) error {
	if len(sourceIDs) == 0 {
		return nil
	}

	var pending []string
	if sourceType == app.TokenSourceTypeToken {
		pending = sourceIDs
	} else {
		ids, err := deleteTokensFromSources(tx, sourceType, sourceIDs)
		if err != nil {
			return err
		}
		pending = ids
	}

	seen := map[string]struct{}{}
	for len(pending) > 0 {
		batch := make([]string, 0, len(pending))
		for _, id := range pending {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				batch = append(batch, id)
			}
		}
		if len(batch) == 0 {
			return nil
		}

		if res := tx.Where("source_token_id = ANY(?)", pq.Array(batch)).Delete(&app.OAuthAuthorizationCode{}); res.Error != nil {
			return errors.Wrap(res.Error, "unable to revoke derived authorization codes")
		}
		if res := tx.Where("source_token_id = ANY(?)", pq.Array(batch)).Delete(&app.OAuthRefreshToken{}); res.Error != nil {
			return errors.Wrap(res.Error, "unable to revoke derived refresh tokens")
		}
		if res := tx.Where("source_token_id = ANY(?)", pq.Array(batch)).Delete(&app.DeviceCode{}); res.Error != nil {
			return errors.Wrap(res.Error, "unable to revoke derived device codes")
		}

		ids, err := deleteTokensFromSources(tx, app.TokenSourceTypeToken, batch)
		if err != nil {
			return err
		}
		pending = ids
	}

	return nil
}

func deleteTokensFromSources(tx *gorm.DB, sourceType app.TokenSourceType, sourceIDs []string) ([]string, error) {
	var ids []string
	if res := tx.Unscoped().Model(&app.Token{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("source_type = ? AND source_id = ANY(?)", sourceType, pq.Array(sourceIDs)).
		Order("id").
		Pluck("id", &ids); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to look up derived tokens")
	}
	if len(ids) == 0 {
		return nil, nil
	}

	if res := tx.Where("id = ANY(?)", pq.Array(ids)).Delete(&app.Token{}); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to revoke derived tokens")
	}

	return ids, nil
}
