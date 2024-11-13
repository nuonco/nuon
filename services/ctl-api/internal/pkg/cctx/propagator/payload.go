package propagator

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type Payload struct {
	OrgID     string         `json:"org_id"`
	AccountID string         `json:"account_id"`
	LogStream *app.LogStream `json:"log_stream"`
}

func FetchPayload(ctx context.Context) (*Payload, error) {
	acctID, err := cctx.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	logStream, _ := cctx.GetLogStreamContext(ctx)

	return &Payload{
		OrgID:     orgID,
		AccountID: acctID,
		LogStream: logStream,
	}, nil
}
