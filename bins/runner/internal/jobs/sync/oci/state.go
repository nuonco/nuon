package containerimage

import (
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

type handlerState struct {
	// state for an individual run, that can not be reused
	plan       *plantypes.SyncOCIPlan
	descriptor *ocispec.Descriptor

	jobID          string
	jobExecutionID string
}
