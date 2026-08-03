package permissions

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const componentBuildsObjectSuffix = ":component_builds"

func ComponentBuildsObject(orgID string) string {
	return orgID + componentBuildsObjectSuffix
}

func RequestObject(ctx *gin.Context, orgID string) string {
	if ctx.Request.Method != http.MethodPost {
		return orgID
	}

	switch ctx.FullPath() {
	case "/v1/apps/:app_id/components/:component_id/builds", "/v1/components/:component_id/builds":
		return ComponentBuildsObject(orgID)
	default:
		return orgID
	}
}

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
