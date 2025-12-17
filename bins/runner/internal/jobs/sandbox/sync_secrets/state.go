package terraform

import (
	"time"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/workspace"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/types/outputs"
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
