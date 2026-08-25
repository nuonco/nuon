package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func TestRequireWriteScope(t *testing.T) {
	// read-only token is blocked
	ctx := keys.WithTokenRole(context.Background(), string(app.RoleTypeOrgReadOnly))
	assert.Error(t, requireWriteScope(ctx))

	// support and admin scopes allowed
	for _, role := range []app.RoleType{app.RoleTypeOrgSupport, app.RoleTypeOrgAdmin} {
		ctx := keys.WithTokenRole(context.Background(), string(role))
		assert.NoError(t, requireWriteScope(ctx), "role %s should allow writes", role)
	}

	// no role set (non-OAuth token) is not restricted here
	assert.NoError(t, requireWriteScope(context.Background()))
}
