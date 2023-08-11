package hooks

import (
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// Return a middleware that sets the hook handlers as parameters on the context.
func New(orgHooks app.OrgHooks, appHooks app.AppHooks, installHooks app.InstallHooks) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(app.OrgHooksKey, orgHooks)
		c.Set(app.AppHooksKey, appHooks)
		c.Set(app.InstallHooksKey, installHooks)

		c.Next()
	}
}
