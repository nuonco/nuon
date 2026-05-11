// Package activities exposes Temporal activities for the composite_errors
// domain — the Workflow-side wrappers that allow workflow code to record a
// CompositeError without depending on the database directly.
//
// The Helpers package is the single source of write logic; activities are a
// thin RPC boundary around it.
package activities

import (
	"go.uber.org/fx"

	cehelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/composite_errors/helpers"
)

type Params struct {
	fx.In

	Helpers *cehelpers.Helpers
}

type Activities struct {
	helpers *cehelpers.Helpers
}

func New(params Params) *Activities {
	return &Activities{
		helpers: params.Helpers,
	}
}
