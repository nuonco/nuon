package account

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

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

		if res := tx.Where("source_token_id IN ?", batch).Delete(&app.OAuthAuthorizationCode{}); res.Error != nil {
			return errors.Wrap(res.Error, "unable to revoke derived authorization codes")
		}
		if res := tx.Where("source_token_id IN ?", batch).Delete(&app.OAuthRefreshToken{}); res.Error != nil {
			return errors.Wrap(res.Error, "unable to revoke derived refresh tokens")
		}
		if res := tx.Where("source_token_id IN ?", batch).Delete(&app.DeviceCode{}); res.Error != nil {
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
		Where("source_type = ? AND source_id IN ?", sourceType, sourceIDs).
		Order("id").
		Pluck("id", &ids); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to look up derived tokens")
	}
	if len(ids) == 0 {
		return nil, nil
	}

	if res := tx.Where("id IN ?", ids).Delete(&app.Token{}); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to revoke derived tokens")
	}

	return ids, nil
}
