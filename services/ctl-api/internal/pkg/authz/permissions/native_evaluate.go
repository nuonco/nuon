package permissions

import "slices"

type NativeTarget struct {
	ResolvedTarget
	RelatedJobs []Scope
}

type NativePermissionCheck struct {
	API        API
	Permission PermissionID
	Target     NativeTarget
	Grants     []*CompiledPolicy
	Ceiling    []*CompiledPolicy
}

const ReasonInvalidAPI MatchReason = "invalid_api"

func (c *Catalog) Match(input NativePermissionCheck) PermissionMatch {
	if _, err := ParseAPI(string(input.API)); err != nil {
		return PermissionMatch{Reason: ReasonInvalidAPI}
	}
	definition, ok := c.definitions[input.Permission]
	if !ok || len(definition.Members) != 0 {
		return PermissionMatch{Reason: ReasonInvalidPermission}
	}
	if err := input.Target.ValidateShape(); err != nil {
		return PermissionMatch{Reason: ReasonInvalidTarget}
	}
	if !slices.Contains(definition.Targets, input.Target.Kind) {
		return PermissionMatch{Reason: ReasonInvalidTarget}
	}
	for _, job := range input.Target.RelatedJobs {
		if input.Target.Kind != KindLogStream || job.Kind != KindRunnerJob || job.OrgID != input.Target.OrgID || job.ValidateShape() != nil {
			return PermissionMatch{Reason: ReasonInvalidTarget}
		}
	}
	grants, err := c.operationScopes(input.Grants, input.Target.OrgID, input.API, input.Permission)
	if err != nil {
		return PermissionMatch{Reason: ReasonInvalidGrant}
	}
	ceiling, err := c.operationScopes(input.Ceiling, input.Target.OrgID, input.API, input.Permission)
	if err != nil {
		return PermissionMatch{Reason: ReasonInvalidCeiling}
	}
	if !nativeScopesPermit(grants, input) {
		return PermissionMatch{Reason: ReasonNoGrant}
	}
	if input.Ceiling != nil && !nativeScopesPermit(ceiling, input) {
		return PermissionMatch{Reason: ReasonCeilingExceeded}
	}
	return PermissionMatch{Matched: true, Reason: ReasonMatched}
}

func nativeScopesPermit(scopes []NativeScope, input NativePermissionCheck) bool {
	for _, scope := range scopes {
		if scope.covers(input.Target, input.Permission) {
			return true
		}
	}
	return false
}

func (s NativeScope) covers(target NativeTarget, operation PermissionID) bool {
	var rootID string
	switch s.Kind {
	case KindOrg:
		rootID = target.OrgID
	case KindApp:
		rootID = target.AppID
	case KindInstall:
		rootID = target.InstallID
	case KindRunnerJob:
		if target.Kind == KindLogStream && operation == LogStreamRead {
			for _, job := range target.RelatedJobs {
				if s.selects(job.ID) {
					return true
				}
			}
			return false
		}
		if target.Kind == KindRunnerJob {
			rootID = target.ID
		}
	default:
		if s.Kind == target.Kind {
			rootID = target.ID
		}
	}
	return rootID != "" && s.selects(rootID)
}

func (s NativeScope) selects(id string) bool {
	return slices.Contains(s.IDs, "*") || slices.Contains(s.IDs, id)
}

func nativeScopeSupports(scope, target ResourceKind, operation PermissionID) bool {
	switch scope {
	case KindOrg:
		return true
	case KindApp:
		return target != KindOrg
	case KindInstall:
		return target != KindOrg && target != KindApp
	case KindRunnerJob:
		return target == KindRunnerJob || (target == KindLogStream && operation == LogStreamRead)
	default:
		return scope == target
	}
}
