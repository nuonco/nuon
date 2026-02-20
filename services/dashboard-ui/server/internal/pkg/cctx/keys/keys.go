package keys

// Context keys for the BFF server. Mirrors the ctl-api cctx/keys pattern.
const (
	AccountIDKey  string = "account_id"
	OrgIDKey      string = "org_id"
	IsEmployeeKey string = "is_employee"
	TokenKey      string = "token"
	MetricsKey    string = "metrics"
	TraceIDKey    string = "trace_id"
	APIClientKey  string = "api_client"
)
