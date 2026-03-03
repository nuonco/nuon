package keys

import "context"

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
)

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
