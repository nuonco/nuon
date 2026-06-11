package plan

import (
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
)

// installComponentOverride looks up a per-component install-level override value
// from the install inputs (stored under a reserved synthetic input name) and
// renders it against the install state so it can reference {{.nuon.*}}.
//
// It returns "" when no override is set, which callers treat as an exact no-op.
func (p *Planner) installComponentOverride(
	st *state.State,
	stateData map[string]any,
	inputName string,
) (string, error) {
	if st == nil || st.Inputs == nil {
		return "", nil
	}

	raw, ok := st.Inputs.Inputs[inputName]
	if !ok || raw == "" {
		return "", nil
	}

	rendered, err := render.RenderV2(raw, stateData)
	if err != nil {
		return "", errors.Wrap(err, "unable to render component override")
	}
	return rendered, nil
}
