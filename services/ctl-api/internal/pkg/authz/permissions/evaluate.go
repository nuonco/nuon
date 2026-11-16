package permissions

import "slices"

type Target struct {
	ResolvedTarget
	RelatedJobs []Scope
}

type Check struct {
	API        API
	Permission PermissionID
	Target     Target
	Grants     []*CompiledPolicy
	Ceiling    []*CompiledPolicy
}

const ReasonInvalidAPI MatchReason = "invalid_api"

func (c *Catalog) Match(input Check) PermissionMatch {
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
	if !scopesPermit(grants, input) {
		return PermissionMatch{Reason: ReasonNoGrant}
	}
	if input.Ceiling != nil && !scopesPermit(ceiling, input) {
		return PermissionMatch{Reason: ReasonCeilingExceeded}
	}
	return PermissionMatch{Matched: true, Reason: ReasonMatched}
}

func scopesPermit(scopes []GrantScope, input Check) bool {
	for _, scope := range scopes {
		if scope.covers(input.Target, input.Permission) {
			return true
		}
	}
	return false
}

func (s GrantScope) covers(target Target, operation PermissionID) bool {
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

func (s GrantScope) selects(id string) bool {
	return slices.Contains(s.IDs, "*") || slices.Contains(s.IDs, id)
}

func scopeSupports(scope, target ResourceKind, operation PermissionID) bool {
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
