package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

const unforgottenEmitterReason = "disabled because install was unforgotten"

type AdminUnforgetInstallRequest struct{}

// @ID						AdminUnforgetInstall
// @Summary				unforget an install
// @Description.markdown	unforget_install.md
// @Param					install_id	path	string							true	"install ID"
// @Param					req			body	AdminUnforgetInstallRequest	true	"Input"
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean}	true
// @Router					/v1/installs/{install_id}/admin-unforget [POST]
func (s *service) AdminUnforgetInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var install app.Install
	err := s.db.WithContext(ctx).
		Unscoped().
		Scopes(scopes.WithDisableViews).
		Where(app.Install{ID: installID}).
		First(&install).Error
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}
	if install.DeletedAt == 0 {
		ctx.Error(stderr.ErrConflict{
			Err:         fmt.Errorf("install %s is not forgotten", installID),
			Description: "install is not forgotten",
		})
		return
	}
	cctx.SetOrgIDGinContext(ctx, install.OrgID)

	if err := s.unforgetInstall(ctx, &install); err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, true)
}

func (s *service) unforgetInstall(ctx context.Context, install *app.Install) error {
	forgottenAt := install.DeletedAt
	installOwnerType := plugins.TableName(s.db, app.Install{})
	runnerOwnerType := plugins.TableName(s.db, app.Runner{})

	var stackIDs []string
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conflict app.Install
		err := tx.Scopes(scopes.WithDisableViews).
			Where(app.Install{AppID: install.AppID, Name: install.Name}).
			Where("id <> ?", install.ID).
			First(&conflict).Error
		if err == nil {
			return stderr.ErrConflict{
				Err:         fmt.Errorf("install name %s is in use", install.Name),
				Description: "another install already uses this name",
			}
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("unable to check install name: %w", err)
		}

		stackIDs, err = pluckIDs(tx, &app.InstallStack{}, forgottenAt, app.InstallStack{InstallID: install.ID}, true)
		if err != nil {
			return fmt.Errorf("unable to list install stacks: %w", err)
		}
		runnerGroupIDs, err := pluckIDs(tx, &app.RunnerGroup{}, forgottenAt, app.RunnerGroup{
			OwnerID:   install.ID,
			OwnerType: installOwnerType,
		}, false)
		if err != nil {
			return fmt.Errorf("unable to list runner groups: %w", err)
		}
		var runnerIDs []string
		if len(runnerGroupIDs) > 0 {
			runnerIDs, err = pluckIDs(tx, &app.Runner{}, forgottenAt, "runner_group_id IN ?", false, runnerGroupIDs)
			if err != nil {
				return fmt.Errorf("unable to list runners: %w", err)
			}
		}
		queueIDs, err := pluckIDs(tx, &app.Queue{}, forgottenAt, app.Queue{
			OwnerID:   install.ID,
			OwnerType: installOwnerType,
		}, false)
		if err != nil {
			return fmt.Errorf("unable to list install queues: %w", err)
		}
		if len(runnerIDs) > 0 {
			runnerQueueIDs, err := pluckIDs(tx, &app.Queue{}, forgottenAt, "owner_id IN ? AND owner_type = ?", false, runnerIDs, runnerOwnerType)
			if err != nil {
				return fmt.Errorf("unable to list runner queues: %w", err)
			}
			queueIDs = append(queueIDs, runnerQueueIDs...)
		}

		if err := restoreInstallAssociations(ctx, tx, install, forgottenAt); err != nil {
			return err
		}
		if len(runnerGroupIDs) > 0 {
			if err := restoreForgottenWhere(tx, &app.Runner{}, forgottenAt, "runner_group_id IN ?", runnerGroupIDs); err != nil {
				return fmt.Errorf("unable to restore runners: %w", err)
			}
		}
		if len(runnerIDs) > 0 {
			if err := restoreForgottenWhere(tx, &app.Queue{}, forgottenAt, "owner_id IN ? AND owner_type = ?", runnerIDs, runnerOwnerType); err != nil {
				return fmt.Errorf("unable to restore runner queues: %w", err)
			}
		}
		if len(queueIDs) > 0 {
			if err := restoreEmittersDisabled(ctx, tx, forgottenAt, queueIDs); err != nil {
				return err
			}
		}
		if err := restoreStackServiceAccounts(tx, stackIDs); err != nil {
			return err
		}

		install.Labels.RemoveKeys([]string{
			"nuon.co/forgotten-by-email",
			"nuon.co/forgotten-by-account-id",
		})
		res := tx.Unscoped().Scopes(scopes.WithDisableViews).
			Model(&app.Install{}).
			Where(app.Install{ID: install.ID}).
			Updates(map[string]any{
				"deleted_at": 0,
				"labels":     install.Labels,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to restore install: %w", res.Error)
		}
		if res.RowsAffected < 1 {
			return fmt.Errorf("install not found %s: %w", install.ID, gorm.ErrRecordNotFound)
		}
		return ensureUnforgottenStackServiceAccounts(ctx, tx, install.OrgID, install.ID, stackIDs)
	})
	return err
}

func ensureUnforgottenStackServiceAccounts(ctx context.Context, tx *gorm.DB, orgID, installID string, stackIDs []string) error {
	acctClient := account.New(account.Params{DB: tx})
	authzClient := authz.New(authz.Params{DB: tx})
	ownerType := plugins.TableName(tx, app.InstallStack{})
	for _, stackID := range stackIDs {
		binding, err := acctClient.EnsureManagedServiceAccount(ctx, account.ManagedServiceAccountRequest{
			OrgID:       orgID,
			OwnerType:   ownerType,
			OwnerID:     stackID,
			Purpose:     app.ManagedServiceAccountPurposeStack,
			InstanceKey: "default",
			Subject:     stackID,
		})
		if err != nil {
			return fmt.Errorf("unable to ensure stack service account: %w", err)
		}
		if err := authzClient.EnsureStackInstallRole(ctx, orgID, installID, binding.AccountID); err != nil {
			return fmt.Errorf("unable to ensure stack service account role: %w", err)
		}
		if err := authzClient.RemoveAccountOrgRoleByType(ctx, app.RoleTypeOrgAdmin, orgID, binding.AccountID); err != nil {
			return fmt.Errorf("unable to revoke stack service account org admin: %w", err)
		}
	}
	return nil
}

func pluckIDs(tx *gorm.DB, model any, forgottenAt soft_delete.DeletedAt, where any, includeLive bool, args ...any) ([]string, error) {
	var ids []string
	query := tx.Unscoped().Scopes(scopes.WithDisableViews).Model(model).Where(where, args...)
	if includeLive {
		query = query.Where("(deleted_at = 0 OR deleted_at >= ?)", forgottenAt)
	} else {
		query = query.Where("deleted_at >= ?", forgottenAt)
	}
	res := query.Pluck("id", &ids)
	return ids, res.Error
}

func restoreInstallAssociations(ctx context.Context, tx *gorm.DB, install *app.Install, forgottenAt soft_delete.DeletedAt) error {
	stmt := &gorm.Statement{DB: tx}
	if err := stmt.Parse(install); err != nil {
		return fmt.Errorf("unable to parse install schema: %w", err)
	}

	rels := append([]*schema.Relationship{}, stmt.Schema.Relationships.HasOne...)
	rels = append(rels, stmt.Schema.Relationships.HasMany...)
	for _, rel := range rels {
		if rel.FieldSchema == nil || rel.FieldSchema.LookUpField("DeletedAt") == nil {
			continue
		}
		conds := rel.ToQueryConditions(ctx, reflect.ValueOf(install))
		if len(conds) == 0 || associationINEmpty(conds) {
			return fmt.Errorf("unable to restore %s: install id is missing", rel.Name)
		}

		model := reflect.New(rel.FieldSchema.ModelType).Interface()
		q := tx.Unscoped().Scopes(scopes.WithDisableViews).Model(model)
		for _, cond := range conds {
			q = q.Where(cond)
		}
		res := q.Where("deleted_at >= ?", forgottenAt).Update("deleted_at", 0)
		if res.Error != nil {
			return fmt.Errorf("unable to restore %s: %w", rel.Name, res.Error)
		}
	}
	return nil
}

func associationINEmpty(conds []clause.Expression) bool {
	for _, cond := range conds {
		in, ok := cond.(clause.IN)
		if ok && len(in.Values) == 0 {
			return true
		}
	}
	return false
}

func restoreForgottenWhere(tx *gorm.DB, model any, forgottenAt soft_delete.DeletedAt, query string, args ...any) error {
	res := tx.Unscoped().Scopes(scopes.WithDisableViews).
		Model(model).
		Where(query, args...).
		Where("deleted_at >= ?", forgottenAt).
		Update("deleted_at", 0)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func restoreEmittersDisabled(ctx context.Context, tx *gorm.DB, forgottenAt soft_delete.DeletedAt, queueIDs []string) error {
	disabled := app.NewCompositeStatus(ctx, app.StatusDisabled)
	disabled.StatusHumanDescription = unforgottenEmitterReason
	res := tx.Unscoped().Scopes(scopes.WithDisableViews).
		Model(&app.QueueEmitter{}).
		Where("queue_id IN ?", queueIDs).
		Where("deleted_at >= ?", forgottenAt).
		Updates(map[string]any{
			"deleted_at": 0,
			"status":     &disabled,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to restore queue emitters: %w", res.Error)
	}
	return nil
}

func restoreStackServiceAccounts(tx *gorm.DB, stackIDs []string) error {
	for _, stackID := range stackIDs {
		accountID, err := restoreStackServiceAccount(tx, stackID)
		if err != nil {
			return err
		}
		if accountID == "" {
			continue
		}
		if err := restoreManagedStackBinding(tx, accountID, stackID); err != nil {
			return err
		}
	}
	return nil
}

func restoreStackServiceAccount(tx *gorm.DB, stackID string) (string, error) {
	var live app.Account
	err := tx.Scopes(scopes.WithDisableViews).
		Where(app.Account{Subject: stackID, AccountType: app.AccountTypeService}).
		First(&live).Error
	if err == nil {
		return live.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("unable to look up stack service account: %w", err)
	}

	var acct app.Account
	err = tx.Unscoped().Scopes(scopes.WithDisableViews).
		Where(app.Account{Subject: stackID, AccountType: app.AccountTypeService}).
		Where("deleted_at > ?", 0).
		Order("deleted_at DESC").
		First(&acct).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("unable to look up forgotten stack service account: %w", err)
	}
	res := tx.Unscoped().Scopes(scopes.WithDisableViews).
		Model(&app.Account{}).
		Where(app.Account{ID: acct.ID}).
		Update("deleted_at", 0)
	if res.Error != nil {
		return "", fmt.Errorf("unable to restore stack service account: %w", res.Error)
	}
	return acct.ID, nil
}

func restoreManagedStackBinding(tx *gorm.DB, accountID, stackID string) error {
	res := tx.Unscoped().
		Model(&app.ManagedServiceAccount{}).
		Where(app.ManagedServiceAccount{
			AccountID:   accountID,
			OwnerType:   plugins.TableName(tx, app.InstallStack{}),
			OwnerID:     stackID,
			Purpose:     app.ManagedServiceAccountPurposeStack,
			InstanceKey: "default",
		}).
		Where("deleted_at > ?", 0).
		Update("deleted_at", 0)
	if res.Error != nil {
		return fmt.Errorf("unable to restore stack service account binding: %w", res.Error)
	}
	return nil
}
