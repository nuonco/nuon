package org

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// An account carrying org ids but no loaded orgs used to index past the end of
// Orgs. The panic escaped the recovery middleware, which is registered after this
// one, so the caller got a proxy 503 with no body rather than an error.
func TestRunnerMiddlewareOrgsShorterThanOrgIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	acct := &app.Account{ID: "acc_one"}
	acct.OrgIDs = []string{"org_one"}

	rr := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/stacks/inl_one/config", nil)
	cctx.SetAccountGinContext(ctx, acct)

	m := &runnerMiddleware{}
	require.NotPanics(t, func() { m.Handler()(ctx) })

	assert.True(t, ctx.IsAborted())
	require.NotEmpty(t, ctx.Errors)
	assert.Contains(t, ctx.Errors[0].Error(), "org", "must fail on the missing org, not on a missing account")
}
