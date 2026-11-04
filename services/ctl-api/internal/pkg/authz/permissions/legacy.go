package permissions

import (
	"fmt"
	"strings"
)

type PermissionSource struct {
	RoleID     string
	PolicyID   string
	Permission Permission
	Grant      *ScopedPermissions
}

type AccountPermissions struct {
	Sources map[string]PermissionSource
}

func LegacySources(grants []ScopedPermissions) (map[string]PermissionSource, error) {
	sources := make(map[string]PermissionSource)
	for i := range grants {
		grant := &grants[i]
		if err := grant.Scope.ValidateShape(); err != nil {
			return nil, err
		}
		add := func(object string, verb Permission, set PermissionSet) error {
			if _, exists := sources[object]; exists {
				return fmt.Errorf("multiple rules target legacy object %q", object)
			}
			sources[object] = PermissionSource{
				Permission: verb, Grant: &ScopedPermissions{Scope: grant.Scope, Permissions: set},
			}
			return nil
		}
		if grant.Scope.Kind == KindOrg {
			verb, err := LegacyOrgPermission(grant.Permissions)
			if err != nil {
				return nil, err
			}
			if verb != PermissionUnknown {
				if err := add(grant.Scope.OrgID, verb, grant.Permissions); err != nil {
					return nil, err
				}
			}
			continue
		}
		for permission := range grant.Permissions {
			var kind ResourceKind
			verb := PermissionAll
			switch {
			case grant.Scope.Kind == KindInstall && permission == InstallAll:
				kind = KindInstall
			case grant.Scope.Kind == KindInstall && permission == TelemetryAll:
				kind = KindTelemetry
			case grant.Scope.Kind == KindInstall && permission == StackAll:
				kind = KindStack
			case grant.Scope.Kind == KindRunnerJob && permission == RunnerJobAll:
				kind = KindRunnerJob
			case grant.Scope.Kind == KindPhoneHome && permission == PhoneHomeCreate:
				kind, verb = KindPhoneHome, PermissionCreate
			default:
				return nil, fmt.Errorf("permission %q at %q scope is not supported by the legacy operation adapter", permission, grant.Scope.Kind)
			}
			if err := add(Object(grant.Scope.OrgID, kind, grant.Scope.ID), verb, PermissionSet{permission: {}}); err != nil {
				return nil, err
			}
		}
	}
	return sources, nil
}

func (p *AccountPermissions) Add(object string, source PermissionSource) {
	if previous, ok := p.Sources[object]; ok && previous.Permission == PermissionAll {
		return
	}
	p.Sources[object] = source
}

func (p *AccountPermissions) Grants() []ScopedPermissions {
	grants := make([]ScopedPermissions, 0)
	for _, source := range p.Sources {
		if source.Grant != nil {
			grants = append(grants, *source.Grant)
		}
	}
	return grants
}

func (p *AccountPermissions) Legacy() Set {
	legacy := Set(NewSet())
	for object, source := range p.Sources {
		if source.Grant == nil {
			legacy[object] = source.Permission
		}
	}
	return legacy
}

func (p *AccountPermissions) CanPerform(object string, verb Permission, readOnly bool) error {
	denied := NoAccessError{ObjectID: object, Permission: verb}
	if p == nil || (readOnly && verb != PermissionRead) {
		return denied
	}
	if err := p.Legacy().CanPerform(object, verb); err == nil {
		return nil
	}
	if source, ok := p.Sources[object]; ok && source.Grant != nil && source.Grant.Scope.Kind != KindOrg {
		if source.Permission == PermissionAll || source.Permission == verb {
			return nil
		}
	}
	orgID, _, _ := strings.Cut(object, ":")
	required := OrgWrite
	switch verb {
	case PermissionRead:
		required = OrgRead
	case PermissionCreate, PermissionUpdate, PermissionDelete:
	default:
		source := p.Sources[orgID]
		if source.Grant == nil {
			return denied
		}
		full, err := LegacyOrgPermission(source.Grant.Permissions)
		if err != nil || full != PermissionAll {
			return denied
		}
	}
	grants := p.Grants()
	ceiling := grants
	target := Scope{Kind: KindOrg, OrgID: orgID, ID: orgID}
	if readOnly {
		ceiling = []ScopedPermissions{{Scope: target, Permissions: PermissionSet{OrgRead: {}}}}
	}
	if MatchPermissions(PermissionCheck{
		Permission: required,
		Target:     ResolvedTarget{Scope: target},
		Grants:     grants,
		Ceiling:    ceiling,
	}).Matched {
		return nil
	}
	return denied
}

func LegacyOrgPermission(set PermissionSet) (Permission, error) {
	if len(set) == 0 {
		return PermissionUnknown, nil
	}
	if _, ok := set[OrgRead]; ok {
		if len(set) == 1 {
			return PermissionRead, nil
		}
		if _, ok := set[OrgWrite]; ok && len(set) == 2 {
			return PermissionAll, nil
		}
	}
	return PermissionUnknown, fmt.Errorf("bundle is not supported by the legacy operation adapter")
}
