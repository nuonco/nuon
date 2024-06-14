package authz

import (
	"github.com/gin-gonic/gin"

	authcontext "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/auth/context"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz/permissions"
)

func CanCreate(ctx *gin.Context, objectID string) error {
	acct, err := authcontext.FromContext(ctx)
	if err != nil {
		return err
	}

	return acct.AllPermissions.CanPerform(objectID, permissions.PermissionCreate)
}

func CanRead(ctx *gin.Context, objectID string) error {
	acct, err := authcontext.FromContext(ctx)
	if err != nil {
		return err
	}

	return acct.AllPermissions.CanPerform(objectID, permissions.PermissionRead)
}

func CanUpdate(ctx *gin.Context, objectID string) error {
	acct, err := authcontext.FromContext(ctx)
	if err != nil {
		return err
	}

	return acct.AllPermissions.CanPerform(objectID, permissions.PermissionUpdate)
}

func CanDelete(ctx *gin.Context, objectID string) error {
	acct, err := authcontext.FromContext(ctx)
	if err != nil {
		return err
	}

	return acct.AllPermissions.CanPerform(objectID, permissions.PermissionDelete)
}
