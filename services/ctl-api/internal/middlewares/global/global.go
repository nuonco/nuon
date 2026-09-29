package global

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

var globalEndpointList map[[2]string]struct{} = map[[2]string]struct{}{
	{"POST", "/v1/orgs"}:                                                  {},
	{"GET", "/v1/orgs"}:                                                   {},
	{"POST", "/v1/general/metrics"}:                                       {},
	{"GET", "/v1/general/current-user"}:                                   {},
	{"GET", "/v1/sandboxes"}:                                              {},
	{"GET", "/v1/sandboxes/:sandbox_id"}:                                  {},
	{"GET", "/v1/sandboxes/:sandbox_id/releases"}:                         {},
	{"POST", "/v1/general/waitlist"}:                                      {},
	{"POST", "/v1/installs/:install_id/phone-home/:phone_home_id"}:        {},
	{"GET", "/v1/account"}:                                                {},
	{"GET", "/v1/account/user-journeys"}:                                  {},
	{"POST", "/v1/account/user-journeys"}:                                 {},
	{"PATCH", "/v1/account/user-journeys/:journey_name/steps/:step_name"}: {},
	{"POST", "/v1/account/user-journeys/:journey_name/reset"}:             {},
	{"POST", "/v1/account/user-journeys/:journey_name/complete"}:          {},
	{"GET", "/v1/auth/me"}:                                                {},
	{"GET", "/v1/auth/validate"}:                                          {},

	{"GET", "/v1/onboarding/example-apps"}:                {},
	{"POST", "/v1/onboarding"}:                            {},
	{"GET", "/v1/onboarding/current"}:                     {},
	{"POST", "/v1/onboarding/current/steps/organization"}: {},
	{"DELETE", "/v1/onboarding/current"}:                  {},
}

type middleware struct {
	l *zap.Logger
}

func (m middleware) Name() string {
	return "global"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method
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
