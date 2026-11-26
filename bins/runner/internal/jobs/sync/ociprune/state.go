package ociprune

import (
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

type handlerState struct {
	plan   *plantypes.PruneOCIPlan
	result *Result

	jobID          string
	jobExecutionID string
}
