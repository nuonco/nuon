package authz

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type NativeAuthorityInput struct {
	AccountID    string
	CredentialID string
	OrgID        string
	API          permissions.API
	Ceiling      []byte
}

type NativeAuthority struct {
	accountID    string
	credentialID string
	orgID        string
	api          permissions.API
	catalog      *permissions.Catalog
	grants       []*permissions.CompiledPolicy
	ceiling      []*permissions.CompiledPolicy
}

func (c *Client) LoadNativeAuthority(ctx context.Context, catalog *permissions.Catalog, input NativeAuthorityInput) (*NativeAuthority, error) {
	if input.AccountID == "" || input.CredentialID == "" {
		return nil, fmt.Errorf("native authority requires authenticated account and credential IDs")
	}
	if _, err := permissions.ParseAPI(string(input.API)); err != nil {
		return nil, err
	}
	if len(input.Ceiling) == 0 {
		return nil, fmt.Errorf("native credential requires an explicit ceiling document")
	}
	document, err := catalog.DecodePolicy(input.Ceiling)
	if err != nil {
		return nil, fmt.Errorf("decode native credential ceiling: %w", err)
	}
	ceiling, err := catalog.Compile(document, input.OrgID)
	if err != nil {
		return nil, fmt.Errorf("compile native credential ceiling: %w", err)
	}
	grants, err := c.LoadNativeAccountPermissions(ctx, input.AccountID, input.OrgID, catalog)
	if err != nil {
		return nil, err
	}
	return &NativeAuthority{
		accountID: input.AccountID, credentialID: input.CredentialID, orgID: input.OrgID,
		api: input.API, catalog: catalog, grants: grants, ceiling: []*permissions.CompiledPolicy{ceiling},
	}, nil
}

func (a *NativeAuthority) ValidateIdentity(accountID, credentialID, orgID string, api permissions.API) error {
	if a == nil || a.catalog == nil || a.ceiling == nil ||
		a.accountID != accountID || a.credentialID != credentialID || a.orgID != orgID || a.api != api {
		return fmt.Errorf("native authority does not match the authenticated request")
	}
	return nil
}

func (a *NativeAuthority) Check(operation permissions.PermissionID, target permissions.NativeTarget) permissions.PermissionMatch {
	if a == nil || a.catalog == nil || a.ceiling == nil {
		return permissions.PermissionMatch{Reason: permissions.ReasonInvalidCeiling}
	}
	if target.OrgID != a.orgID {
		return permissions.PermissionMatch{Reason: permissions.ReasonInvalidTarget}
	}
	return a.catalog.Match(permissions.NativePermissionCheck{
		API: a.api, Permission: operation, Target: target, Grants: a.grants, Ceiling: a.ceiling,
	})
}

func (a *NativeAuthority) AuthorizedIDs(ctx context.Context, operation permissions.PermissionID, resources *NativeResources) (*gorm.DB, error) {
	if a == nil || a.catalog == nil || a.ceiling == nil {
		return nil, fmt.Errorf("native authority requires a credential ceiling")
	}
	return resources.AuthorizedIDs(ctx, a.catalog, permissions.NativePermissionQuery{
		API: a.api, Permission: operation, Kind: resources.kind, OrgID: a.orgID, Grants: a.grants, Ceiling: a.ceiling,
	})
}
