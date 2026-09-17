package propagator

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type Payload struct {
	OrgID        string         `json:"org_id"`
	AccountID    string         `json:"account_id"`
	TraceID      string         `json:"trace_id"`
	WorkflowType string         `json:"workflow_type,omitempty"`
	OrgName      string         `json:"org_name,omitempty"`
	InstallName  string         `json:"install_name,omitempty"`
	LogStream    *app.LogStream `json:"log_stream,omitempty"`
}

func FetchPayload(ctx cctx.ValueContext) (*Payload, error) {
	acctID, _ := cctx.AccountIDFromContext(ctx)
	orgID, _ := cctx.OrgIDFromContext(ctx)
	traceID := cctx.TraceIDFromContext(ctx)
	logStream, _ := cctx.GetLogStreamContext(ctx)

	return &Payload{
		OrgID:        orgID,
		AccountID:    acctID,
		TraceID:      traceID,
		WorkflowType: cctx.WorkflowTypeFromContext(ctx),
		OrgName:      cctx.OrgNameFromContext(ctx),
		InstallName:  cctx.InstallNameFromContext(ctx),
		LogStream:    logStream,
	}, nil
}
