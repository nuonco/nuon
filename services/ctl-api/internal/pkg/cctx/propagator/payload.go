package propagator

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type Payload struct {
	OrgID           string                      `json:"org_id"`
	AccountID       string                      `json:"account_id"`
	LogStream       *app.LogStream              `json:"log_stream,omitempty"`
	InstallWorkflow *app.InstallWorkflowContext `json:"install_workflow,omitempty"`
}

func FetchPayload(ctx cctx.ValueContext) (*Payload, error) {
	acctID, _ := cctx.AccountIDFromContext(ctx)
	orgID, _ := cctx.OrgIDFromContext(ctx)
	logStream, _ := cctx.GetLogStreamContext(ctx)
	workflow, _ := cctx.GetInstallWorkflowContext(ctx)

	return &Payload{
		OrgID:           orgID,
		AccountID:       acctID,
		LogStream:       logStream,
		InstallWorkflow: workflow,
	}, nil
}
