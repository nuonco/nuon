package app

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// Terminal step errors related to component builds. Centralised here so all
// call sites that fail because a deployable build is unavailable share the
// same reason_code and user-facing copy. Step error handling short-circuits
// auto-retry when these markers are present (see signal.IsTerminalError).

// ErrNoComponentBuild returns a terminal error for the case where no build
// exists for a component. componentID is used in the user-facing message.
func ErrNoComponentBuild(componentID string) error {
	return signal.NewTerminalError(
		"no_component_build",
		"No active build found for component (id %s). Ensure there is an active build for the component before retrying.",
		componentID,
	)
}

// ErrComponentBuildErrored returns a terminal error for the case where the
// latest build of a component is in the error state. buildID is used in the
// user-facing message.
func ErrComponentBuildErrored(buildID string) error {
	return signal.NewTerminalError(
		"component_build_errored",
		"Component build (id %s) is in an error state. Ensure there is an active build for the component before retrying.",
		buildID,
	)
}
