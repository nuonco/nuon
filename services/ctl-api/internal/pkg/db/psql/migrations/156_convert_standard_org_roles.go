package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (m *Migrations) Migration156ConvertStandardOrgRoles(ctx context.Context, db *gorm.DB) error {
	counts := make(map[string]int)
	defer func() {
		m.l.Info("standard org role conversion results", zap.Any("counts", counts))
	}()
	lastID := ""
	for {
		var roles []app.Role
		if err := db.WithContext(ctx).Model(&app.Role{}).Select("id").
			Where("id > ?", lastID).Order("id").Limit(100).Scan(&roles).Error; err != nil {
			counts["failed"]++
			return err
		}
		if len(roles) == 0 {
			return nil
		}
		for _, role := range roles {
			counts["scanned"]++
			var classification authz.OrgRoleClassification
			err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				source, err := loadStandardOrgRoleSnapshot(tx, role.ID)
				if err != nil {
					return err
				}
				classification, err = authz.ClassifyOrgRole(source)
				if err != nil || classification.Status != "eligible" {
					return err
				}
				counts["eligible"]++
				var api permissions.API
				if classification.LegacyAPI != nil {
					api = *classification.LegacyAPI
				}
				bundle, err := permissions.RolePolicy(permissions.RoleName(source.Role.RoleType), api)
				if err != nil {
					return err
				}
				document, err := bundle.MarshalJSON()
				if err != nil {
					return err
				}
				policy := source.Role.Policies[0]
				res := tx.Model(&app.Policy{}).
					Where(app.Policy{ID: policy.ID, RoleID: source.Role.ID, OrgID: policy.OrgID}).
					Where("document IS NULL AND permissions = ?", policy.Permissions).
					UpdateColumns(map[string]any{"document": document, "permissions": nil})
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected != 1 {
					return fmt.Errorf("policy %s changed during conversion", policy.ID)
				}
				res = tx.Model(&app.Role{}).
					Where(app.Role{ID: source.Role.ID, OrgID: source.Role.OrgID, RoleType: source.Role.RoleType, Version: source.Role.Version}).
					UpdateColumn("version", app.RoleVersionV3)
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected != 1 {
					return fmt.Errorf("role %s changed during conversion", source.Role.ID)
				}
				return nil
			}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
			if err != nil {
				counts["failed"]++
				return fmt.Errorf("convert standard role %s: %w", role.ID, err)
			}
			switch classification.Status {
			case "eligible":
				counts["converted"]++
			case "already_v3":
				counts["already_v3"]++
			case "skipped":
				counts["skipped"]++
				counts["skipped_"+classification.Reason]++
				m.l.Debug("standard org role conversion skipped", zap.String("role_id", role.ID), zap.String("reason", classification.Reason))
			}
		}
		lastID = roles[len(roles)-1].ID
	}
}

func loadStandardOrgRoleSnapshot(db *gorm.DB, roleID string) (authz.OrgRoleSnapshot, error) {
	var source authz.OrgRoleSnapshot
	if err := db.Model(&app.Role{}).Where(app.Role{ID: roleID}).
		Clauses(clause.Locking{Strength: "UPDATE"}).Scan(&source.Role).Error; err != nil {
		return source, err
	}
	if source.Role.ID == "" {
		return source, nil
	}
	if err := db.Model(&app.Policy{}).Where(app.Policy{RoleID: roleID}).
		Clauses(clause.Locking{Strength: "UPDATE"}).Scan(&source.Role.Policies).Error; err != nil {
		return source, err
	}
	var private int64
	if err := db.Unscoped().Model(&app.ManagedServiceAccount{}).
		Where(app.ManagedServiceAccount{PrivateRoleID: &roleID}).Count(&private).Error; err != nil {
		return source, err
	}
	source.Private = private > 0
	if !source.Role.HasPolicyDocument() && (source.Private ||
		(source.Role.RoleType != app.RoleTypeOrgAdmin && source.Role.RoleType != app.RoleTypeOrgReadOnly)) {
		return source, nil
	}
	source.Role.Org = &app.Org{}
	if err := db.Unscoped().Model(&app.Org{}).Select("id", "deleted_at").
		Where(&app.Org{ID: source.Role.OrgID.String}, "id").Scan(source.Role.Org).Error; err != nil {
		return source, err
	}
	if err := db.Model(&app.Role{}).Where(&app.Role{OrgID: source.Role.OrgID}, "org_id").Scan(&source.OrgRoles).Error; err != nil {
		return source, err
	}
	if err := db.Model(&app.AccountRole{}).Where(app.AccountRole{RoleID: roleID}).Scan(&source.Bindings).Error; err != nil {
		return source, err
	}
	for i := range source.Bindings {
		binding := &source.Bindings[i]
		if err := db.Unscoped().Model(&app.Account{}).Select("id", "deleted_at").
			Where(&app.Account{ID: binding.AccountID}, "id").Scan(&binding.Account).Error; err != nil {
			return source, err
		}
		if err := db.Model(&app.Role{}).
			Joins("JOIN account_roles ON account_roles.role_id = roles.id AND account_roles.deleted_at = 0").
			Where("account_roles.account_id = ?", binding.AccountID).
			Scan(&binding.Account.Roles).Error; err != nil {
			return source, err
		}
		for j := range binding.Account.Roles {
			role := &binding.Account.Roles[j]
			if err := db.Model(&app.Policy{}).Where(app.Policy{RoleID: role.ID}).Scan(&role.Policies).Error; err != nil {
				return source, err
			}
		}
	}
	return source, nil
}
