package global

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// globalEndpointList is a list of endpoints that are not scoped to an org,
// but still need to be authenticated.
var globalEndpointList map[[2]string]struct{} = map[[2]string]struct{}{
	{"POST", "/v1/orgs"}:                                           {},
	{"GET", "/v1/orgs"}:                                            {},
	{"POST", "/v1/general/metrics"}:                                {},
	{"GET", "/v1/general/current-user"}:                            {},
	{"GET", "/v1/sandboxes"}:                                       {},
	{"GET", "/v1/sandboxes/:sandbox_id"}:                           {},
	{"GET", "/v1/sandboxes/:sandbox_id/releases"}:                  {},
	{"POST", "/v1/general/waitlist"}:                               {},
	{"POST", "/v1/installs/:install_id/phone-home/:phone_home_id"}: {},
}

type middleware struct {
	l *zap.Logger
}

func (m middleware) Name() string {
	return "global"
}

// Handler marks a request as "global" if it's being sent to one of the endpoints listed in globalEndpointList.
func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method
		// full path will return the _matched_ path, such as `/v1/sandboxes/:id`
		path := ctx.FullPath()

		key := [2]string{
			method,
			path,
		}
		_, found := globalEndpointList[key]
		if found {
			m.l.Debug("marking request as global", zap.String("endpoint", fmt.Sprintf("%s:%s", method, path)))
		}

		cctx.SetIsGlobal(ctx, found)
	}
}

func New(l *zap.Logger) *middleware {
	return &middleware{
		l: l,
	}
}
