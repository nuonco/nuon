package authz

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

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
