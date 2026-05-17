package apisyncer

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
)

// syncRunbook syncs a runbook definition and its config via the API.
// TODO(runbooks): Wire up SDK client methods once the API endpoints are created
// and the SDK is regenerated. The SDK methods needed are:
//   - GetAppRunbook(ctx, appID, name)
//   - CreateRunbook(ctx, appID, req)
//   - UpdateRunbook(ctx, runbookID, req)
//   - CreateRunbookConfig(ctx, runbookID, req)
func (s *syncer) syncRunbook(ctx context.Context, resource string, runbook *config.RunbookConfig) (string, string, error) {
	_ = runbook
	return "", "", sync.SyncInternalErr{
		Description: fmt.Sprintf("runbook sync not yet implemented for API syncer (runbook: %s)", runbook.Name),
		Err:         fmt.Errorf("SDK methods not yet generated"),
	}
}
