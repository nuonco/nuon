package propagator

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type Payload struct {
	OrgID             string                 `json:"org_id" temporaljson:"org_id"`
	AccountID         string                 `json:"account_id" temporaljson:"account_id"`
	TraceID           string                 `json:"trace_id" temporaljson:"trace_id"`
	WorkflowTelemetry cctx.WorkflowTelemetry `json:"workflow_telemetry,omitempty" temporaljson:"workflow_telemetry,omitempty"`
	LogStream         *app.LogStream         `json:"log_stream,omitempty" temporaljson:"log_stream,omitempty"`
	QueueID           string                 `json:"queue_id,omitempty" temporaljson:"queue_id,omitempty"`
}

// legacyPayload matches headers written before Payload carried temporaljson
// tags, when the converter keyed fields by Go name. Long-running workflows
// (cron emitters) still carry those headers.
type legacyPayload struct {
	OrgID             string                 `temporaljson:"OrgID"`
	AccountID         string                 `temporaljson:"AccountID"`
	TraceID           string                 `temporaljson:"TraceID"`
	WorkflowTelemetry cctx.WorkflowTelemetry `temporaljson:"WorkflowTelemetry"`
	LogStream         *app.LogStream         `temporaljson:"LogStream"`
}

func (p *Payload) fillFromLegacy(l legacyPayload) {
	if p.OrgID == "" {
		p.OrgID = l.OrgID
	}
	if p.AccountID == "" {
		p.AccountID = l.AccountID
	}
	if p.TraceID == "" {
		p.TraceID = l.TraceID
	}
	if p.LogStream == nil {
		p.LogStream = l.LogStream
	}
	p.WorkflowTelemetry = l.WorkflowTelemetry.Merge(p.WorkflowTelemetry)
}

func FetchPayload(ctx cctx.ValueContext) (*Payload, error) {
	acctID, _ := cctx.AccountIDFromContext(ctx)
	orgID, _ := cctx.OrgIDFromContext(ctx)
	traceID := cctx.TraceIDFromContext(ctx)
	logStream, _ := cctx.GetLogStreamContext(ctx)

	return &Payload{
		OrgID:             orgID,
		AccountID:         acctID,
		TraceID:           traceID,
		WorkflowTelemetry: cctx.WorkflowTelemetryFromContext(ctx),
		LogStream:         logStream,
		QueueID:           cctx.QueueIDFromContext(ctx),
	}, nil
}
