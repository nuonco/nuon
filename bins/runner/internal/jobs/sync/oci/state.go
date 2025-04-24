package containerimage

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

type handlerState struct {
	// state for an individual run, that can not be reused
	plan      *plantypes.SyncOCIPlan
	workspace workspace.Workspace

	jobID          string
	jobExecutionID string
}
