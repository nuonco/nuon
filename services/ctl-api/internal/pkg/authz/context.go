package authz

import (
	"fmt"

	"github.com/gin-gonic/gin"

	authcontext "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz/permissions"
)

func CanCreate(ctx *gin.Context, objectID string) error {
	acct, err := authcontext.FromGinContext(ctx)
	if err != nil {
		return err
	}

	if objectID == "" {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	return acct.AllPermissions.CanPerform(objectID, permissions.PermissionCreate)
}

func CanRead(ctx *gin.Context, objectID string) error {
	acct, err := authcontext.FromGinContext(ctx)
	if err != nil {
		return err
	}

	if objectID == "" {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	return acct.AllPermissions.CanPerform(objectID, permissions.PermissionRead)
}

func CanUpdate(ctx *gin.Context, objectID string) error {
	acct, err := authcontext.FromGinContext(ctx)
	if err != nil {
		return err
	}

	if objectID == "" {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	return acct.AllPermissions.CanPerform(objectID, permissions.PermissionUpdate)
}

func CanDelete(ctx *gin.Context, objectID string) error {
	acct, err := authcontext.FromGinContext(ctx)
	if err != nil {
		return err
	}

	if objectID == "" {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	return acct.AllPermissions.CanPerform(objectID, permissions.PermissionDelete)
}
