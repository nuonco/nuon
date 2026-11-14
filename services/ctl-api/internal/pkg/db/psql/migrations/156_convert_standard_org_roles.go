package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

func (m *Migrations) Migration156ConvertStandardOrgRoles(ctx context.Context, db *gorm.DB) error {
	counts := make(map[string]int)
	defer func() {
		m.l.Info("compatible role conversion results", zap.Any("counts", counts))
	}()
	standardTypes := []app.RoleType{app.RoleTypeOrgAdmin, app.RoleTypeOrgReadOnly, app.RoleTypeOrgSupport, app.RoleTypeInstaller, app.RoleTypeRunner}
	scopedTypes := []app.RoleType{app.RoleTypeRunnerInstall, app.RoleTypeRunnerJob, app.RoleTypeStack, app.RoleTypeTelemetryCollector, app.RoleTypePhoneHome, app.RoleTypeLogStream}
	legacyPolicies := db.Model(&app.Policy{}).Select("role_id").Where("document IS NULL")
	lastID := ""
	for {
		var roles []app.Role
		if err := db.WithContext(ctx).Model(&app.Role{}).Select("id").
			Where("version IN ?", []int{app.RoleVersionV1, app.RoleVersionV2}).
			Where("role_type IN ? OR role_type IN ?", standardTypes, scopedTypes).
			Where("id IN (?)", legacyPolicies).
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
			var err error
			for attempt := 0; attempt < 4; attempt++ {
				err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
					source, err := loadStandardOrgRoleSnapshot(tx, role.ID)
					if err != nil {
						return err
					}
					classification, err = authz.ClassifyOrgRole(source)
					if err != nil || classification.Status != "eligible" {
						return err
					}
					document, err := classification.Policy.MarshalJSON()
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
				var pgErr *pgconn.PgError
				if attempt == 3 || !errors.As(err, &pgErr) || (pgErr.Code != "40001" && pgErr.Code != "40P01") {
					break
				}
				m.l.Warn("retrying compatible role conversion", zap.String("role_id", role.ID), zap.Int("attempt", attempt+1), zap.Error(err))
				timer := time.NewTimer(100 * time.Millisecond * time.Duration(1<<attempt))
				select {
				case <-ctx.Done():
					timer.Stop()
				case <-timer.C:
				}
				if ctx.Err() != nil {
					err = ctx.Err()
					break
				}
				counts["retried"]++
			}
			if err != nil {
				counts["failed"]++
				return fmt.Errorf("convert compatible role %s: %w", role.ID, err)
			}
			switch classification.Status {
			case "eligible":
				counts["eligible"]++
				counts["converted"]++
			case "already_v3":
				counts["already_v3"]++
			case "skipped":
				counts["skipped"]++
				counts["skipped_"+classification.Reason]++
				m.l.Debug("compatible role conversion skipped", zap.String("role_id", role.ID), zap.String("reason", classification.Reason))
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
	source.Role.Org = &app.Org{}
	if err := db.Unscoped().Model(&app.Org{}).Select("id", "deleted_at").
		Where(&app.Org{ID: source.Role.OrgID.String}, "id").Scan(source.Role.Org).Error; err != nil {
		return source, err
	}
	if err := db.Model(&app.AccountRole{}).Where(app.AccountRole{RoleID: roleID}).Scan(&source.Bindings).Error; err != nil {
		return source, err
	}
	for batch := range slices.Chunk(source.Bindings, 500) {
		accountIDs := make([]string, 0, len(batch))
		for _, binding := range batch {
			accountIDs = append(accountIDs, binding.AccountID)
		}
		var accounts []app.Account
		if err := db.Unscoped().Model(&app.Account{}).Select("id", "deleted_at").
			Where("id IN ?", accountIDs).Scan(&accounts).Error; err != nil {
			return source, err
		}
		byID := make(map[string]app.Account, len(accounts))
		for _, account := range accounts {
			byID[account.ID] = account
		}
		for i := range batch {
			batch[i].Account = byID[batch[i].AccountID]
		}
	}
	return source, nil
}
