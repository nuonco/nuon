package authz

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const authorityKey = "nuon.authz.authority"

func SetAuthorityGinContext(ctx *gin.Context, authority *Authority) {
	ctx.Set(authorityKey, authority)
}

func HasAuthority(ctx cctx.ValueContext) bool {
	return ctx.Value(authorityKey) != nil
}

func AuthorityFromContext(ctx cctx.ValueContext) (*Authority, error) {
	authority, ok := ctx.Value(authorityKey).(*Authority)
	if !ok || authority == nil {
		return nil, fmt.Errorf("authority was not set by authentication")
	}
	return authority, nil
}

func CanCreate(ctx *gin.Context, objectID string) error {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return err
	}

	if objectID == "" {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	return acct.Authority.CanPerform(objectID, permissions.PermissionCreate, false)
}

func CanRead(ctx *gin.Context, objectID string) error {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return err
	}

	if objectID == "" {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	return acct.Authority.CanPerform(objectID, permissions.PermissionRead, false)
}

func CanUpdate(ctx *gin.Context, objectID string) error {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return err
	}

	if objectID == "" {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	return acct.Authority.CanPerform(objectID, permissions.PermissionUpdate, false)
}

func CanDelete(ctx *gin.Context, objectID string) error {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return err
	}

	if objectID == "" {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	return acct.Authority.CanPerform(objectID, permissions.PermissionDelete, false)
}
