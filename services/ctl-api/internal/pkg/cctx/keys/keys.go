package keys

import (
	"context"
)

// All of the context keys are defined here so we can use them in different contexts.
//
// While most every package can use the cctx helpers directly, since they do leverage models, anything in the models
// that needs information from the context can not rely on that package directly, otherwise a circular dependency will
// be created.
const (
	AccountCtxKey         string = "account"
	AccountIDCtxKey       string = "account_id"
	BlobServiceCtxKey     string = "blob_service"
	CfgCtxKey             string = "config"
	IsGlobalKey           string = "is_global"
	InstallWorkflowCtxKey string = "workflow"
	FlowCtxKey            string = "flow"
	IsEmployeeCtxKey      string = "is_employee"
	LoggerFieldsCtxKey    string = "logger_fields"
	LogStreamCtxKey       string = "log_stream"
	MetricsKey            string = "metrics"
	OrgCtxKey             string = "org"
	OrgIDCtxKey           string = "org_id"
	OffPaginationCtxKey   string = "offset_pagination"
	IsPublicKey           string = "is_public"
	RunnerCtxKey          string = "runner"
	RunnerIDCtxKey        string = "runner_id"
	DisableViewCtxKey     string = "disable_view"
	PatcherCtxKey         string = "patcher"
	TraceIDCtxKey         string = "trace_id"
	FlowWorkflowIDCtxKey  string = "flow_workflow_id"
	FlowInstallIDCtxKey   string = "flow_install_id"
	OrgSelectorCtxKey     string = "mcp_org_selector"
	TokenRoleCtxKey       string = "token_role"
)

// OrgSelectFunc persists the selected org for the current MCP session. It is
// injected by the MCP server so leaf tool handlers (in other packages) can
// change the session's active org without importing the server package.
type OrgSelectFunc func(orgID string)

// WithOrgSelector attaches an org-selector to the context.
func WithOrgSelector(ctx context.Context, fn OrgSelectFunc) context.Context {
	return context.WithValue(ctx, OrgSelectorCtxKey, fn)
}

// OrgSelectorFromContext returns the org-selector, or nil if none is set (e.g.
// outside the MCP server).
func OrgSelectorFromContext(ctx context.Context) OrgSelectFunc {
	fn, _ := ctx.Value(OrgSelectorCtxKey).(OrgSelectFunc)
	return fn
}

// WithTokenRole attaches the authenticating token's org role/scope to the
// context (e.g. org_read_only, org_support, org_admin).
func WithTokenRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, TokenRoleCtxKey, role)
}

// TokenRoleFromContext returns the authenticating token's role, or "" if unset.
func TokenRoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(TokenRoleCtxKey).(string)
	return role
}

// CreatedByIDFromContext returns the account ID from context.
// Returns empty string if not set. This is safe to call from leaf packages
// that cannot import the full cctx package due to circular dependencies.
func CreatedByIDFromContext(ctx context.Context) string {
	val := ctx.Value(AccountIDCtxKey)
	valStr, ok := val.(string)
	if !ok {
		return ""
	}
	return valStr
}

// OrgIDFromContext returns the org ID from context.
// Returns empty string if not set. This is safe to call from leaf packages
// that cannot import the full cctx package due to circular dependencies.
func OrgIDFromContext(ctx context.Context) string {
	val := ctx.Value(OrgIDCtxKey)
	valStr, ok := val.(string)
	if !ok {
		return ""
	}
	return valStr
}

func FlowWorkflowIDFromContext(ctx context.Context) string {
	s, _ := ctx.Value(FlowWorkflowIDCtxKey).(string)
	return s
}

func FlowInstallIDFromContext(ctx context.Context) string {
	s, _ := ctx.Value(FlowInstallIDCtxKey).(string)
	return s
}
