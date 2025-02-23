package authz

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

func CanCreate(ctx *gin.Context, objectID string) error {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return err
	}

	if objectID == "" {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	return acct.AllPermissions.CanPerform(objectID, permissions.PermissionCreate)
}

func CanRead(ctx *gin.Context, objectID string) error {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return err
	}

	if objectID == "" {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	return acct.AllPermissions.CanPerform(objectID, permissions.PermissionRead)
}

func CanUpdate(ctx *gin.Context, objectID string) error {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return err
	}

	if objectID == "" {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	return acct.AllPermissions.CanPerform(objectID, permissions.PermissionUpdate)
}

func CanDelete(ctx *gin.Context, objectID string) error {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return err
	}

	if objectID == "" {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	return acct.AllPermissions.CanPerform(objectID, permissions.PermissionDelete)
}
