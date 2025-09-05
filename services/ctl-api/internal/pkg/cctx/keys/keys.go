package keys

// All of the context keys are defined here so we can use them in different contexts.
//
// While most every package can use the cctx helpers directly, since they do leverage models, anything in the models
// that needs information from the context can not rely on that package directly, otherwise a circular dependency will
// be created.
const (
	AccountCtxKey         string = "account"
	AccountIDCtxKey       string = "account_id"
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
