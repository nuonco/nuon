package authz

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type Identity struct {
	AccountID       string
	CredentialID    string
	OrgID           string
	CredentialOrgID string
	API             permissions.API
}

type AccessContext struct {
	authority    *Authority
	requirements []Requirement
	selection    *Selection
}

func (a *Authority) Access() *AccessContext {
	return &AccessContext{authority: a}
}

func (a *AccessContext) Identity() Identity {
	return Identity{
		AccountID: a.authority.accountID, CredentialID: a.authority.credentialID,
		OrgID: a.authority.orgID, CredentialOrgID: a.authority.credentialOrgID, API: a.authority.api,
	}
}

func (a *AccessContext) Matches(authority *Authority) bool {
	return a != nil && a.authority != nil && a.authority == authority
}

func (a *AccessContext) Requirements() []Requirement {
	result := slices.Clone(a.requirements)
	for i := range result {
		result[i].Target.RelatedJobs = slices.Clone(result[i].Target.RelatedJobs)
	}
	return result
}

func (a *AccessContext) Target(operation permissions.PermissionID) (permissions.Target, error) {
	var found *permissions.Target
	for _, requirement := range a.Requirements() {
		if requirement.Operation != operation {
			continue
		}
		if found != nil {
			return permissions.Target{}, fmt.Errorf("operation %s has multiple checked targets", operation)
		}
		target := requirement.Target
		found = &target
	}
	if found == nil {
		return permissions.Target{}, fmt.Errorf("operation %s has no checked target", operation)
	}
	return *found, nil
}

func (a *AccessContext) Collection() (*Selection, error) {
	if a.selection == nil {
		return nil, fmt.Errorf("collection requires an authorized ID query")
	}
	return a.selection, nil
}

func (a *AccessContext) Require(requirements ...Requirement) (*AccessContext, error) {
	if err := a.authority.Require(requirements...); err != nil {
		return nil, err
	}
	result := &AccessContext{authority: a.authority, requirements: a.Requirements(), selection: a.selection}
	for _, requirement := range requirements {
		requirement.Target.RelatedJobs = slices.Clone(requirement.Target.RelatedJobs)
		slices.SortFunc(requirement.Target.RelatedJobs, func(left, right permissions.Scope) int {
			return cmp.Compare(left.ID, right.ID)
		})
		requirement.Target.RelatedJobs = slices.Compact(requirement.Target.RelatedJobs)
		duplicate := false
		for _, previous := range result.requirements {
			if previous.Target.Scope != requirement.Target.Scope {
				continue
			}
			if previous.Target.ResolvedTarget != requirement.Target.ResolvedTarget ||
				!slices.Equal(previous.Target.RelatedJobs, requirement.Target.RelatedJobs) {
				return nil, fmt.Errorf("conflicting checked resource %s", requirement.Target.ID)
			}
			if previous.Operation == requirement.Operation {
				duplicate = true
			}
		}
		if !duplicate {
			result.requirements = append(result.requirements, requirement)
		}
	}
	return result, nil
}

func (a *AccessContext) Select(ctx context.Context, operation permissions.PermissionID, resources *Resources) (*AccessContext, error) {
	if a.selection != nil {
		return nil, fmt.Errorf("a route cannot apply multiple collection authorization rules")
	}
	selection, err := a.authority.Select(ctx, operation, resources)
	if err != nil {
		return nil, err
	}
	return &AccessContext{authority: a.authority, requirements: a.Requirements(), selection: selection}, nil
}
