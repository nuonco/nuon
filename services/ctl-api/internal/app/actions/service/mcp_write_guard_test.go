package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func TestRequireWriteScope(t *testing.T) {
	ctx := keys.WithTokenRole(context.Background(), string(app.RoleTypeOrgReadOnly))
	assert.Error(t, requireWriteScope(ctx))

	for _, role := range []app.RoleType{app.RoleTypeOrgSupport, app.RoleTypeOrgAdmin} {
		ctx := keys.WithTokenRole(context.Background(), string(role))
		assert.NoError(t, requireWriteScope(ctx), "role %s should allow writes", role)
	}

	assert.NoError(t, requireWriteScope(context.Background()))
}
