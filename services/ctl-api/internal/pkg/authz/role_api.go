package authz

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func APIFromRole(role *app.Role) (permissions.API, bool, error) {
	if role == nil || role.Version < app.RoleVersionV2 {
		return "", false, nil
	}
	version := role.Version
	if version == app.RoleVersionV3 {
		if err := role.ValidateLegacyCompatibility(); err != nil {
			return "", true, err
		}
		version = role.LegacyVersion
		if version == app.RoleVersionV1 {
			return "", false, nil
		}
	}
	if version != app.RoleVersionV2 {
		return "", true, fmt.Errorf("role format %d is not supported by legacy API authorization", role.Version)
	}
	var found permissions.API
	var ok bool
	for _, policy := range role.Policies {
		api, has := permissions.APIFromPermissions(policy.Permissions)
		if !has {
			continue
		}
		if ok && found != api {
			return "", true, fmt.Errorf("role %s has more than one api", role.ID)
		}
		found = api
		ok = true
	}
	if !ok {
		return "", true, fmt.Errorf("role %s is missing an api permission", role.RoleType)
	}
	return found, true, nil
}

func APIForRoles(roles []app.Role) (permissions.API, bool, error) {
	var found permissions.API
	var enforce bool
	for i := range roles {
		api, v2, err := APIFromRole(&roles[i])
		if err != nil {
			return "", true, err
		}
		if !v2 {
			continue
		}
		if enforce && found != api {
			return "", true, fmt.Errorf("account has v2 roles for more than one api")
		}
		found = api
		enforce = true
	}
	return found, enforce, nil
}

func assertRoleAPICompatible(db *gorm.DB, accountID string, role *app.Role) error {
	if role == nil {
		return fmt.Errorf("legacy assignment requires a role")
	}
	if role.Version == app.RoleVersionV3 {
		if err := role.ValidateLegacyCompatibility(); err != nil {
			return err
		}
	} else {
		if role.Version != app.RoleVersionV1 && role.Version != app.RoleVersionV2 {
			return fmt.Errorf("unsupported legacy assignment format")
		}
		for _, policy := range role.Policies {
			if len(policy.Document) != 0 {
				return fmt.Errorf("legacy assignment cannot use a structured policy")
			}
		}
	}
	api, v2, err := APIFromRole(role)
	if err != nil {
		return err
	}
	var roles []app.Role
	if err := db.Joins("JOIN account_roles ON account_roles.role_id = roles.id AND account_roles.deleted_at = 0").
		Preload("Policies").
		Where("account_roles.account_id = ?", accountID).
		Where("roles.id <> ?", role.ID).
		Find(&roles).Error; err != nil {
		return fmt.Errorf("load account roles: %w", err)
	}
	existing, enforce, err := APIForRoles(roles)
	if err != nil {
		return err
	}
	if v2 && enforce && existing != api {
		return fmt.Errorf("account already has a v2 role for api %s", existing)
	}
	return nil
}
