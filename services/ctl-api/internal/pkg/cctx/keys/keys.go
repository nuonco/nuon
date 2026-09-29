package keys

import (
	"context"
)

const (
	AccountCtxKey           string = "account"
	AccountIDCtxKey         string = "account_id"
	BlobServiceCtxKey       string = "blob_service"
	CfgCtxKey               string = "config"
	IsGlobalKey             string = "is_global"
	InstallWorkflowCtxKey   string = "workflow"
	FlowCtxKey              string = "flow"
	IsEmployeeCtxKey        string = "is_employee"
	LoggerFieldsCtxKey      string = "logger_fields"
	LogStreamCtxKey         string = "log_stream"
	MetricsKey              string = "metrics"
	OrgCtxKey               string = "org"
	OrgIDCtxKey             string = "org_id"
	OffPaginationCtxKey     string = "offset_pagination"
	IsPublicKey             string = "is_public"
	RunnerCtxKey            string = "runner"
	RunnerIDCtxKey          string = "runner_id"
	DisableViewCtxKey       string = "disable_view"
	PatcherCtxKey           string = "patcher"
	TraceIDCtxKey           string = "trace_id"
	FlowWorkflowIDCtxKey    string = "flow_workflow_id"
	FlowInstallIDCtxKey     string = "flow_install_id"
	WorkflowTelemetryCtxKey string = "workflow_telemetry"
	QueueIDCtxKey           string = "queue_id"
	OrgSelectorCtxKey       string = "mcp_org_selector"
	TokenRoleCtxKey         string = "token_role"
)

type WorkflowTelemetry struct {
	OrgID        string `json:"org_id,omitempty"`
	OrgName      string `json:"org_name,omitempty"`
	WorkflowID   string `json:"workflow_id,omitempty"`
	WorkflowType string `json:"workflow_type,omitempty"`
	OwnerID      string `json:"owner_id,omitempty"`
	OwnerType    string `json:"owner_type,omitempty"`
	OwnerName    string `json:"owner_name,omitempty"`
	InstallID    string `json:"install_id,omitempty"`
	InstallName  string `json:"install_name,omitempty"`
}

// why: Merge returns t with every non-empty field of overlay applied, so a caller
// holding partial identity never erases fields resolved further upstream.
func (t WorkflowTelemetry) Merge(overlay WorkflowTelemetry) WorkflowTelemetry {
	if overlay.OrgID != "" {
		t.OrgID = overlay.OrgID
	}
	if overlay.OrgName != "" {
		t.OrgName = overlay.OrgName
	}
	if overlay.WorkflowID != "" {
		t.WorkflowID = overlay.WorkflowID
	}
	if overlay.WorkflowType != "" {
		t.WorkflowType = overlay.WorkflowType
	}
	if overlay.OwnerID != "" {
		t.OwnerID = overlay.OwnerID
	}
	if overlay.OwnerType != "" {
		t.OwnerType = overlay.OwnerType
	}
	if overlay.OwnerName != "" {
		t.OwnerName = overlay.OwnerName
	}
	if overlay.InstallID != "" {
		t.InstallID = overlay.InstallID
	}
	if overlay.InstallName != "" {
		t.InstallName = overlay.InstallName
	}
	return t
}

type OrgSelectFunc func(orgID string)

func WithOrgSelector(ctx context.Context, fn OrgSelectFunc) context.Context {
	return context.WithValue(ctx, OrgSelectorCtxKey, fn)
}

func OrgSelectorFromContext(ctx context.Context) OrgSelectFunc {
	fn, _ := ctx.Value(OrgSelectorCtxKey).(OrgSelectFunc)
	return fn
}

func WithTokenRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, TokenRoleCtxKey, role)
}

func TokenRoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(TokenRoleCtxKey).(string)
	return role
}

func CreatedByIDFromContext(ctx context.Context) string {
	val := ctx.Value(AccountIDCtxKey)
	valStr, ok := val.(string)
	if !ok {
		return ""
	}
	return valStr
}

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
