package authz

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

var ErrDenied = errors.New("authorization denied")

type Requirement struct {
	Operation permissions.PermissionID
	Target    permissions.Target
}

func RequestAuthority(ctx cctx.ValueContext, api permissions.API) (*Authority, error) {
	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if !HasAuthority(ctx) && account.Authority != nil {
		return nil, nil
	}
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	authority, err := AuthorityFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := authority.ValidateIdentity(account.ID, cctx.TokenIDFromContext(ctx), orgID, api); err != nil {
		return nil, err
	}
	return authority, nil
}

type AuthorityInput struct {
	AccountID    string
	CredentialID string
	OrgID        string
	API          permissions.API
	Ceiling      []byte
}

type Authority struct {
	accountID    string
	credentialID string
	orgID        string
	api          permissions.API
	catalog      *permissions.Catalog
	grants       []*permissions.CompiledPolicy
	ceiling      []*permissions.CompiledPolicy
}

func (c *Client) LoadAuthority(ctx context.Context, catalog *permissions.Catalog, input AuthorityInput) (*Authority, error) {
	if input.AccountID == "" || input.CredentialID == "" {
		return nil, fmt.Errorf("authority requires authenticated account and credential IDs")
	}
	if _, err := permissions.ParseAPI(string(input.API)); err != nil {
		return nil, err
	}
	var ceiling []*permissions.CompiledPolicy
	if input.Ceiling != nil {
		document, err := catalog.DecodePolicy(input.Ceiling)
		if err != nil {
			return nil, fmt.Errorf("decode credential ceiling: %w", err)
		}
		compiled, err := catalog.Compile(document, input.OrgID)
		if err != nil {
			return nil, fmt.Errorf("compile credential ceiling: %w", err)
		}
		ceiling = []*permissions.CompiledPolicy{compiled}
	}
	grants, err := c.LoadAccountGrants(ctx, input.AccountID, input.OrgID, catalog)
	if err != nil {
		return nil, err
	}
	return &Authority{
		accountID: input.AccountID, credentialID: input.CredentialID, orgID: input.OrgID,
		api: input.API, catalog: catalog, grants: grants, ceiling: ceiling,
	}, nil
}

func (a *Authority) HasCeiling() bool {
	return a != nil && a.ceiling != nil
}

func (a *Authority) ValidateIdentity(accountID, credentialID, orgID string, api permissions.API) error {
	if a == nil || a.catalog == nil ||
		a.accountID != accountID || a.credentialID != credentialID || a.orgID != orgID || a.api != api {
		return fmt.Errorf("authority does not match the authenticated request")
	}
	return nil
}

func (a *Authority) Check(operation permissions.PermissionID, target permissions.Target) permissions.PermissionMatch {
	if a == nil || a.catalog == nil {
		return permissions.PermissionMatch{Reason: permissions.ReasonInvalidCeiling}
	}
	if target.OrgID != a.orgID {
		return permissions.PermissionMatch{Reason: permissions.ReasonInvalidTarget}
	}
	return a.catalog.Match(permissions.Check{
		API: a.api, Permission: operation, Target: target, Grants: a.grants, Ceiling: a.ceiling,
	})
}

func (a *Authority) Require(requirements ...Requirement) error {
	if len(requirements) == 0 {
		return fmt.Errorf("%w: operation requires an authorization target", ErrDenied)
	}
	for _, requirement := range requirements {
		if match := a.Check(requirement.Operation, requirement.Target); !match.Matched {
			return fmt.Errorf("%w: operation %s denied: %s", ErrDenied, requirement.Operation, match.Reason)
		}
	}
	return nil
}

func (a *Authority) Select(ctx context.Context, operation permissions.PermissionID, resources *Resources) (*Selection, error) {
	if a == nil || a.catalog == nil {
		return nil, fmt.Errorf("authority was not loaded")
	}
	ids, err := resources.AuthorizedIDs(ctx, a.catalog, permissions.Query{
		API: a.api, Permission: operation, Kind: resources.kind, OrgID: a.orgID, Grants: a.grants, Ceiling: a.ceiling,
	})
	if err != nil {
		return nil, err
	}
	return &Selection{ids: ids}, nil
}
