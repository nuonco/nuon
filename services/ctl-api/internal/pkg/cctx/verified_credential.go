package cctx

import (
	"context"
	"maps"
	"slices"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type CredentialSurface string

const (
	SurfacePublicAPI CredentialSurface = "public-api"
	SurfaceMCP       CredentialSurface = "mcp"
)

type CredentialProfile string

const (
	ProfileHuman          CredentialProfile = "human"
	ProfileServiceAccount CredentialProfile = "service-account"
)

type VerifiedCredential struct {
	AccountID string
	TokenID   string
	Surface   CredentialSurface
	Profile   CredentialProfile
	Audience  permissions.API
	Ceiling   []permissions.ScopedPermissions
}

func SetVerifiedCredentialContext(ctx context.Context, credential VerifiedCredential) context.Context {
	return context.WithValue(ctx, keys.VerifiedCredentialKey, credential.clone())
}

func VerifiedCredentialFromContext(ctx ValueContext) (VerifiedCredential, bool) {
	credential, ok := ctx.Value(keys.VerifiedCredentialKey).(VerifiedCredential)
	return credential.clone(), ok
}

func (c VerifiedCredential) clone() VerifiedCredential {
	c.Ceiling = slices.Clone(c.Ceiling)
	for i := range c.Ceiling {
		c.Ceiling[i].Permissions = maps.Clone(c.Ceiling[i].Permissions)
	}
	return c
}
