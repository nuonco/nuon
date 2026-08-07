package permissions

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func FromRequest(ctx *gin.Context) Permission {
	method := strings.ToLower(ctx.Request.Method)

	switch method {
	case "get", "head", "":
		return PermissionRead
	case "delete":
		return PermissionDelete
	case "put", "patch":
		return PermissionUpdate
	case "post":
		return PermissionCreate
	}

	return PermissionUnknown
}
