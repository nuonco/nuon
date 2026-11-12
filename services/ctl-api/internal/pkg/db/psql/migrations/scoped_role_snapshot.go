package migrations

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func loadScopedRoleResource(db *gorm.DB, source *authz.OrgRoleSnapshot) error {
	if len(source.Bindings) != 1 {
		return nil
	}
	policy, err := authz.ScopedRolePolicy(&source.Role)
	if err != nil {
		return nil
	}
	orgID := source.Role.OrgID.String
	grants, err := policy.Compile(orgID, permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID})
	if err != nil {
		return err
	}
	target := grants[0].Scope
	account := source.Bindings[0].Account
	resource := &authz.ScopedRoleSnapshot{}
	source.Resource = resource
	if err := db.Unscoped().Model(&app.ManagedServiceAccount{}).
		Where(&app.ManagedServiceAccount{AccountID: account.ID}, "account_id").Scan(&resource.Owner).Error; err != nil {
		return err
	}
	installID := ""
	if target.Kind == permissions.KindInstall {
		installID = target.ID
	}
	switch source.Role.RoleType {
	case app.RoleTypeLogStream:
		if err := db.Unscoped().Model(&app.LogStream{}).Select("id", "org_id", "deleted_at").
			Where(&app.LogStream{ID: target.ID}, "id").Scan(&resource.LogStream).Error; err != nil {
			return err
		}
	case app.RoleTypeRunnerInstall, app.RoleTypeRunnerJob:
		jobID := target.ID
		if source.Role.RoleType == app.RoleTypeRunnerInstall {
			jobID = account.Subject
		}
		if err := db.Unscoped().Model(&app.RunnerJob{}).Scopes(scopes.WithDisableViews).
			Select("id", "org_id", "deleted_at", "runner_id").Where(&app.RunnerJob{ID: jobID}, "id").Scan(&resource.Job).Error; err != nil {
			return err
		}
		if source.Role.RoleType == app.RoleTypeRunnerInstall {
			runnerID := account.Subject
			if resource.Job.ID != "" {
				runnerID = resource.Job.RunnerID
			}
			if err := db.Unscoped().Model(&app.Runner{}).Select("id", "org_id", "deleted_at", "runner_group_id").
				Where(&app.Runner{ID: runnerID}, "id").Scan(&resource.Runner).Error; err != nil {
				return err
			}
			if err := db.Unscoped().Model(&app.RunnerGroup{}).Select("id", "org_id", "deleted_at", "type", "owner_id", "owner_type").
				Where(&app.RunnerGroup{ID: resource.Runner.RunnerGroupID}, "id").Scan(&resource.RunnerGroup).Error; err != nil {
				return err
			}
		}
	case app.RoleTypeStack:
		if err := db.Unscoped().Model(&app.InstallStack{}).Select("id", "org_id", "deleted_at", "install_id").
			Where(&app.InstallStack{ID: account.Subject}, "id").Scan(&resource.Stack).Error; err != nil {
			return err
		}
	case app.RoleTypePhoneHome:
		if err := db.Unscoped().Model(&app.InstallStackVersion{}).Select("id", "org_id", "deleted_at", "install_id", "install_stack_id").
			Where(&app.InstallStackVersion{ID: target.ID}, "id").Scan(&resource.StackVersion).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Model(&app.InstallStack{}).Select("id", "org_id", "deleted_at", "install_id").
			Where(&app.InstallStack{ID: resource.StackVersion.InstallStackID}, "id").Scan(&resource.Stack).Error; err != nil {
			return err
		}
		installID = resource.StackVersion.InstallID
	}
	if installID != "" {
		if err := db.Unscoped().Model(&app.Install{}).Scopes(scopes.WithDisableViews).
			Select("id", "org_id", "deleted_at").Where(app.Install{ID: installID}).Scan(&resource.Install).Error; err != nil {
			return err
		}
	}
	return nil
}
