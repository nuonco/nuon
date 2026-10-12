package authz

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func RequireUserManaged(tx *gorm.DB, accountID string) error {
	var acct struct{ ID string }
	if err := tx.Model(&app.Account{}).Select("id").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(app.Account{ID: accountID}).Take(&acct).Error; err != nil {
		return fmt.Errorf("lock account for management: %w", err)
	}
	var count int64
	if err := tx.Model(&app.ManagedServiceAccount{}).
		Where(app.ManagedServiceAccount{AccountID: accountID}).Count(&count).Error; err != nil {
		return fmt.Errorf("load managed service account ownership: %w", err)
	}
	if count != 0 {
		return stderr.ErrAuthorization{
			Err:         fmt.Errorf("service account is managed by its owning resource"),
			Description: "manage this service account through its owning resource",
		}
	}
	return nil
}
