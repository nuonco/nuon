package state

import "github.com/powertoolsdev/mono/pkg/types/state"

func (w *Workflows) mapLegacyFields(is *state.State) {
	// NOTE(JM): this is purely for historical and legacy reasons, and will be removed once we migrate all users to
	// the flattened structure
	is.Install = &state.InstallState{
		Populated: true,
		ID:        is.ID,
		Name:      is.Name,
		Sandbox:   *is.Sandbox,
	}
	if is.Domain != nil {
		is.Install.PublicDomain = is.Domain.PublicDomain
		is.Install.InternalDomain = is.Domain.InternalDomain
	}

	if is.Inputs != nil {
		is.Install.Inputs = is.Inputs.Inputs
	}
}
