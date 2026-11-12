package authz

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type ScopedRoleSnapshot struct {
	Owner        app.ManagedServiceAccount
	Install      app.Install
	Stack        app.InstallStack
	StackVersion app.InstallStackVersion
	Job          app.RunnerJob
	Runner       app.Runner
	RunnerGroup  app.RunnerGroup
	LogStream    app.LogStream
}

func scopedRoleOwnership(source OrgRoleSnapshot, scope permissions.Scope) string {
	role, resource := source.Role, source.Resource
	if role.Managed || len(role.Contexts) != 0 || len(source.Bindings) != 1 {
		return "ambiguous_scoped_role"
	}
	account := source.Bindings[0].Account
	if account.AccountType != app.AccountTypeService || resource == nil {
		return "missing_scoped_ownership"
	}
	owner := resource.Owner
	if owner.AccountID != "" && (owner.AccountID != account.ID || owner.OrgID != scope.OrgID || owner.DeletedAt != 0) {
		return "managed_owner_mismatch"
	}
	if source.Private && (owner.PrivateRoleID == nil || *owner.PrivateRoleID != role.ID) {
		return "private_role_owner_mismatch"
	}
	if role.RoleType != app.RoleTypeRunnerInstall && owner.PrivateRoleID != nil && *owner.PrivateRoleID != role.ID {
		return "private_role_owner_mismatch"
	}
	if scope.Kind == permissions.KindInstall && (resource.Install.ID != scope.ID || resource.Install.OrgID != scope.OrgID || resource.Install.DeletedAt != 0) {
		return "missing_or_deleted_install"
	}
	switch role.RoleType {
	case app.RoleTypeLogStream:
		stream := resource.LogStream
		if stream.ID != scope.ID || stream.OrgID != scope.OrgID || stream.DeletedAt != 0 || account.Subject != scope.ID {
			return "log_stream_owner_mismatch"
		}
		if owner.AccountID != "" {
			return "managed_owner_mismatch"
		}
	case app.RoleTypeRunnerInstall:
		if source.Private {
			return "unexpected_private_install_role"
		}
		runnerID := account.Subject
		if resource.Job.ID != "" {
			if resource.Job.ID != account.Subject || resource.Job.OrgID != scope.OrgID || resource.Job.DeletedAt != 0 {
				return "runner_job_owner_mismatch"
			}
			runnerID = resource.Job.RunnerID
		}
		if owner.AccountID != "" && (resource.Job.ID == "" || owner.Purpose != app.ManagedServiceAccountPurposeRunnerJob || owner.OwnerType != "runner_jobs" || owner.OwnerID != resource.Job.ID) {
			return "managed_owner_mismatch"
		}
		if runnerID == "" || resource.Runner.ID != runnerID || resource.Runner.OrgID != scope.OrgID || resource.Runner.DeletedAt != 0 {
			return "runner_owner_mismatch"
		}
		group := resource.RunnerGroup
		if group.ID == "" || group.ID != resource.Runner.RunnerGroupID || group.OrgID != scope.OrgID || group.DeletedAt != 0 || group.Type != app.RunnerGroupTypeInstall || group.OwnerType != "installs" || group.OwnerID != scope.ID {
			return "runner_install_owner_mismatch"
		}
	case app.RoleTypeRunnerJob:
		if resource.Job.ID != scope.ID || resource.Job.OrgID != scope.OrgID || resource.Job.DeletedAt != 0 || account.Subject != scope.ID {
			return "runner_job_owner_mismatch"
		}
		if owner.AccountID != "" && (owner.Purpose != app.ManagedServiceAccountPurposeRunnerJob || owner.OwnerType != "runner_jobs" || owner.OwnerID != scope.ID) {
			return "managed_owner_mismatch"
		}
	case app.RoleTypeStack:
		stack := resource.Stack
		if stack.ID == "" || stack.ID != account.Subject || stack.OrgID != scope.OrgID || stack.DeletedAt != 0 || stack.InstallID != scope.ID {
			return "stack_owner_mismatch"
		}
		if owner.AccountID != "" && (owner.Purpose != app.ManagedServiceAccountPurposeStack || owner.OwnerType != "install_stacks" || owner.OwnerID != stack.ID) {
			return "managed_owner_mismatch"
		}
	case app.RoleTypeTelemetryCollector:
		if owner.AccountID == "" || owner.Purpose != app.ManagedServiceAccountPurposeTelemetryCollector || owner.OwnerType != "installs" || owner.OwnerID != scope.ID {
			return "managed_owner_mismatch"
		}
	case app.RoleTypePhoneHome:
		version, stack, install := resource.StackVersion, resource.Stack, resource.Install
		if version.ID != scope.ID || version.OrgID != scope.OrgID || version.DeletedAt != 0 || account.Subject != scope.ID {
			return "phone_home_version_mismatch"
		}
		if stack.ID == "" || version.InstallStackID != stack.ID || stack.OrgID != scope.OrgID || stack.DeletedAt != 0 || version.InstallID != stack.InstallID {
			return "stack_owner_mismatch"
		}
		if install.ID == "" || install.ID != version.InstallID || install.OrgID != scope.OrgID || install.DeletedAt != 0 {
			return "missing_or_deleted_install"
		}
		if owner.AccountID == "" || owner.Purpose != app.ManagedServiceAccountPurposePhoneHome || owner.OwnerType != "install_stacks" || owner.OwnerID != stack.ID || owner.InstanceKey != scope.ID {
			return "managed_owner_mismatch"
		}
	}
	return ""
}
