package authz

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func lockLegacyRoleSubject(tx *gorm.DB, accountID string) error {
	if _, err := lockRoleSubject(tx, accountID); err != nil {
		return err
	}
	var bindings []app.AccountRole
	if err := tx.Preload("Role").Where(app.AccountRole{AccountID: accountID}).Find(&bindings).Error; err != nil {
		return err
	}
	for _, binding := range bindings {
		if binding.ScopeKind != "" || binding.ScopeID != "" || binding.Role.Version < app.RoleVersionV1 || binding.Role.Version > app.RoleVersionV2 {
			return fmt.Errorf("legacy role mutation cannot change scoped or unsupported bindings")
		}
	}
	return nil
}
