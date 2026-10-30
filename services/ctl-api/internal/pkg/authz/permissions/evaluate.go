package permissions

type ScopedPermissions struct {
	Scope       Scope
	Permissions PermissionSet
}

type PermissionCheck struct {
	Permission PermissionID
	Target     ResolvedTarget
	Grants     []ScopedPermissions
	Ceiling    []ScopedPermissions
}

type MatchReason string

const (
	ReasonMatched           MatchReason = "matched"
	ReasonInvalidPermission MatchReason = "invalid_permission"
	ReasonInvalidTarget     MatchReason = "invalid_target"
	ReasonInvalidGrant      MatchReason = "invalid_grant"
	ReasonInvalidCeiling    MatchReason = "invalid_ceiling"
	ReasonNoGrant           MatchReason = "no_grant"
	ReasonCeilingExceeded   MatchReason = "ceiling_exceeded"
)

type PermissionMatch struct {
	Matched bool
	Reason  MatchReason
}

func MatchPermissions(input PermissionCheck) PermissionMatch {
	kind := input.Permission.resourceKind()
	if kind == "" {
		return PermissionMatch{Reason: ReasonInvalidPermission}
	}
	if err := input.Target.ValidateShape(); err != nil {
		return PermissionMatch{Reason: ReasonInvalidTarget}
	}
	if !input.Permission.appliesTo(input.Target.Kind) {
		return PermissionMatch{Reason: ReasonInvalidTarget}
	}
	if !validGrants(input.Grants) {
		return PermissionMatch{Reason: ReasonInvalidGrant}
	}
	if !validGrants(input.Ceiling) {
		return PermissionMatch{Reason: ReasonInvalidCeiling}
	}
	if !grantsPermit(input.Grants, input.Permission, input.Target) {
		return PermissionMatch{Reason: ReasonNoGrant}
	}
	if !grantsPermit(input.Ceiling, input.Permission, input.Target) {
		return PermissionMatch{Reason: ReasonCeilingExceeded}
	}
	return PermissionMatch{Matched: true, Reason: ReasonMatched}
}

func validGrants(grants []ScopedPermissions) bool {
	for _, grant := range grants {
		if err := grant.Scope.ValidateShape(); err != nil {
			return false
		}
		for permission := range grant.Permissions {
			if permission.resourceKind() == "" {
				return false
			}
		}
	}
	return true
}

func grantsPermit(grants []ScopedPermissions, permission PermissionID, target ResolvedTarget) bool {
	for _, grant := range grants {
		if _, ok := grant.Permissions[permission]; ok && grant.Scope.covers(target) {
			return true
		}
	}
	return false
}
