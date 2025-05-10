package terraform

import (
	"time"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/types/outputs"
)

const (
	defaultFileType string = "file/terraform"
)

type handlerState struct {
	workspace workspace.Workspace

	timeout time.Duration

	// fields set by the plugin execution
	jobExecutionID string
	jobID          string
	plan           *plantypes.SyncSecretsPlan
	outputs        outputs.SyncSecretsOutput
}
