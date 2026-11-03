package permissions

import (
	"encoding/json"
	"fmt"
)

const CatalogRevision = 2

type CustomerRoleName string

const (
	RoleOrgReader     CustomerRoleName = "org_read_only"
	RoleOrgEditor     CustomerRoleName = "org_editor"
	RoleOrgAdmin      CustomerRoleName = "org_admin"
	RoleAppReader     CustomerRoleName = "app_reader"
	RoleAppEditor     CustomerRoleName = "app_editor"
	RoleAppAdmin      CustomerRoleName = "app_admin"
	RoleInstallReader CustomerRoleName = "install_reader"
	RoleInstallEditor CustomerRoleName = "install_editor"
	RoleInstallAdmin  CustomerRoleName = "install_admin"
)

type RoleDefinition struct {
	Name        CustomerRoleName
	Revision    int
	ScopeKind   ResourceKind
	Permissions PermissionSet
}

func (r RoleDefinition) MarshalPolicy() ([]byte, error) {
	return json.Marshal(bundleDocument{
		Kind: PolicyKindBundle, API: APICtl, BindingKind: r.ScopeKind, Permissions: r.Permissions.Values(),
	})
}

func CurrentCustomerRole(name CustomerRoleName) (RoleDefinition, error) {
	values := []PermissionID{InstallRead, RunnerRead, RunnerJobRead, TelemetryRead, LogStreamRead}
	var kind ResourceKind
	switch name {
	case RoleInstallReader, RoleInstallEditor, RoleInstallAdmin:
		kind = KindInstall
	case RoleAppReader, RoleAppEditor, RoleAppAdmin:
		kind = KindApp
		values = append(values, AppRead)
	case RoleOrgReader, RoleOrgEditor, RoleOrgAdmin:
		kind = KindOrg
		values = append(values, AppRead, OrgRead)
	default:
		return RoleDefinition{}, fmt.Errorf("invalid customer role %q", name)
	}

	switch name {
	case RoleInstallEditor, RoleInstallAdmin:
		values = append(values, InstallWrite)
	case RoleAppEditor, RoleAppAdmin, RoleOrgEditor, RoleOrgAdmin:
		values = append(values, AppWrite, InstallWrite)
	}

	switch name {
	case RoleInstallAdmin:
		values = append(values, InstallDelete)
	case RoleAppAdmin:
		values = append(values, AppDelete, InstallDelete)
	case RoleOrgAdmin:
		values = append(values, OrgWrite, AppDelete, InstallDelete)
	}

	set, err := NewPermissionSet(values...)
	if err != nil {
		return RoleDefinition{}, err
	}
	return RoleDefinition{Name: name, Revision: CatalogRevision, ScopeKind: kind, Permissions: set}, nil
}

func BindCurrentCustomerRole(name CustomerRoleName, scope Scope) (ScopedPermissions, error) {
	role, err := CurrentCustomerRole(name)
	if err != nil {
		return ScopedPermissions{}, err
	}
	if err := scope.ValidateShape(); err != nil {
		return ScopedPermissions{}, err
	}
	if scope.Kind != role.ScopeKind {
		return ScopedPermissions{}, fmt.Errorf("role %q requires scope kind %q", name, role.ScopeKind)
	}
	return ScopedPermissions{Scope: scope, Permissions: role.Permissions}, nil
}
