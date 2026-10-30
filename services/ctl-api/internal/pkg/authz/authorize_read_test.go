package authz_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func TestAuthorizeReadRequiresSupportedRequestAndVerifiedCredential(t *testing.T) {
	for name, tc := range map[string]struct {
		change func(*authz.ReadRequest, *cctx.VerifiedCredential)
		reason authz.DecisionReason
	}{
		"write is not a read": {func(r *authz.ReadRequest, _ *cctx.VerifiedCredential) {
			r.Permission = permissions.InstallWrite
		}, authz.ReasonUnsupportedRead},
		"telemetry ingestion is not diagnostic read": {func(r *authz.ReadRequest, _ *cctx.VerifiedCredential) {
			r.Permission = permissions.TelemetryWrite
			r.Target.Kind = permissions.KindTelemetry
		}, authz.ReasonUnsupportedRead},
		"config read is not diagnostic read": {func(r *authz.ReadRequest, _ *cctx.VerifiedCredential) {
			r.Permission = permissions.TelemetryConfigRead
		}, authz.ReasonUnsupportedRead},
		"runner config read is not diagnostic read": {func(r *authz.ReadRequest, _ *cctx.VerifiedCredential) {
			r.Permission = permissions.RunnerConfigRead
			r.Target.Kind = permissions.KindRunner
		}, authz.ReasonUnsupportedRead},
		"runner read requires verified credential": {func(r *authz.ReadRequest, c *cctx.VerifiedCredential) {
			r.Permission = permissions.RunnerRead
			r.Target.Kind = permissions.KindRunner
			c.TokenID = ""
		}, authz.ReasonMissingCredential},
		"job config read is not diagnostic read": {func(r *authz.ReadRequest, _ *cctx.VerifiedCredential) {
			r.Permission = permissions.RunnerJobConfigRead
			r.Target.Kind = permissions.KindRunnerJob
		}, authz.ReasonUnsupportedRead},
		"log writes are not diagnostic reads": {func(r *authz.ReadRequest, _ *cctx.VerifiedCredential) {
			r.Permission = permissions.LogStreamWrite
			r.Target.Kind = permissions.KindLogStream
		}, authz.ReasonUnsupportedRead},
		"job read requires verified credential": {func(r *authz.ReadRequest, c *cctx.VerifiedCredential) {
			r.Permission = permissions.RunnerJobRead
			r.Target.Kind = permissions.KindRunnerJob
			c.TokenID = ""
		}, authz.ReasonMissingCredential},
		"log read requires verified credential": {func(r *authz.ReadRequest, c *cctx.VerifiedCredential) {
			r.Permission = permissions.LogStreamRead
			r.Target.Kind = permissions.KindLogStream
			c.TokenID = ""
		}, authz.ReasonMissingCredential},
		"unimplemented ownership family": {func(r *authz.ReadRequest, _ *cctx.VerifiedCredential) {
			r.Permission = permissions.ExecutionStateRead
		}, authz.ReasonUnsupportedRead},
		"unknown permission": {func(r *authz.ReadRequest, _ *cctx.VerifiedCredential) {
			r.Permission = "install.unknown"
		}, authz.ReasonUnsupportedRead},
		"wrong target kind": {func(r *authz.ReadRequest, _ *cctx.VerifiedCredential) {
			r.Target.Kind = permissions.KindApp
		}, authz.ReasonInvalidRequest},
		"no tenant": {func(r *authz.ReadRequest, _ *cctx.VerifiedCredential) {
			r.Target.OrgID = ""
		}, authz.ReasonInvalidRequest},
		"no account": {func(_ *authz.ReadRequest, c *cctx.VerifiedCredential) {
			c.AccountID = ""
		}, authz.ReasonMissingCredential},
		"no token": {func(_ *authz.ReadRequest, c *cctx.VerifiedCredential) {
			c.TokenID = ""
		}, authz.ReasonMissingCredential},
		"wrong audience": {func(_ *authz.ReadRequest, c *cctx.VerifiedCredential) {
			c.Audience = permissions.APIRunner
		}, authz.ReasonCredentialNotAllowed},
		"unknown surface": {func(_ *authz.ReadRequest, c *cctx.VerifiedCredential) {
			c.Surface = "internal-api"
		}, authz.ReasonCredentialNotAllowed},
		"unknown profile": {func(_ *authz.ReadRequest, c *cctx.VerifiedCredential) {
			c.Profile = "runner"
		}, authz.ReasonCredentialNotAllowed},
		"absent ceiling is not unrestricted": {func(_ *authz.ReadRequest, c *cctx.VerifiedCredential) {
			c.Ceiling = nil
		}, authz.ReasonCeilingExceeded},
		"empty ceiling is not unrestricted": {func(_ *authz.ReadRequest, c *cctx.VerifiedCredential) {
			c.Ceiling = []permissions.ScopedPermissions{}
		}, authz.ReasonCeilingExceeded},
	} {
		t.Run(name, func(t *testing.T) {
			scope := permissions.Scope{Kind: permissions.KindInstall, OrgID: "org_acme", ID: "install_one"}
			req := authz.ReadRequest{Permission: permissions.InstallRead, Target: scope}
			credential := cctx.VerifiedCredential{
				AccountID: "account_one", TokenID: "token_one", Surface: cctx.SurfacePublicAPI,
				Profile: cctx.ProfileHuman, Audience: permissions.APICtl,
				Ceiling: []permissions.ScopedPermissions{{Scope: scope, Permissions: permissions.PermissionSet{permissions.InstallRead: {}}}},
			}
			tc.change(&req, &credential)
			ctx := cctx.SetVerifiedCredentialContext(context.Background(), credential)
			decision, err := authz.New(authz.Params{}).AuthorizeRead(ctx, req)
			require.NoError(t, err)
			require.Equal(t, authz.Decision{Reason: tc.reason}, decision)
		})
	}

	for name, ctx := range map[string]context.Context{
		"no authentication": context.Background(),
		"legacy identity is not verified credential context": cctx.SetAccountContext(context.Background(), &app.Account{ID: "account_one"}),
	} {
		t.Run(name, func(t *testing.T) {
			decision, err := authz.New(authz.Params{}).AuthorizeRead(ctx, authz.ReadRequest{
				Permission: permissions.InstallRead,
				Target:     permissions.Scope{Kind: permissions.KindInstall, OrgID: "org_acme", ID: "install_one"},
			})
			require.NoError(t, err)
			require.Equal(t, authz.Decision{Reason: authz.ReasonMissingCredential}, decision)
		})
	}
}
