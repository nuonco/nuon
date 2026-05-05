package stderr_test

import (
	stderrs "errors"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func TestNewStopError_FieldsPropagated(t *testing.T) {
	cause := stderrs.New("underlying boom")
	u := stderr.NewStopError("no_component_build", "no build for component", map[string]string{
		"component_id": "comp_123",
	}, cause)

	assert.Equal(t, stderr.StepDirectiveStop, u.Directive)
	assert.Equal(t, "no_component_build", u.Code)
	assert.Equal(t, "no build for component", u.Description)
	assert.Equal(t, "comp_123", u.Fields["component_id"])
	assert.Same(t, cause, u.Unwrap())
}

func TestNewSkipError_NoCauseSyntheticMessage(t *testing.T) {
	u := stderr.NewSkipError("already_torn_down", "nothing to do", nil, nil)

	assert.Equal(t, stderr.StepDirectiveSkip, u.Directive)
	require.NotNil(t, u.Unwrap())
	assert.Contains(t, u.Error(), "already_torn_down")
	assert.Contains(t, u.Error(), "nothing to do")
}

func TestIsUserError_UnwrapsThroughPkgErrors(t *testing.T) {
	u := stderr.NewStopError("plan_superseded", "plan superseded", nil, nil)
	wrapped := pkgerrors.Wrap(u, "while running validate")

	got, ok := stderr.IsUserError(wrapped)
	require.True(t, ok)
	assert.Equal(t, "plan_superseded", got.Code)
	assert.Equal(t, stderr.StepDirectiveStop, got.Directive)
}

func TestIsUserError_NotUserError(t *testing.T) {
	_, ok := stderr.IsUserError(stderrs.New("plain error"))
	assert.False(t, ok)
}

func TestIsUserError_NilError(t *testing.T) {
	_, ok := stderr.IsUserError(nil)
	assert.False(t, ok)
}

func TestDirectiveOf(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want stderr.StepDirective
	}{
		{"nil", nil, stderr.StepDirectiveDefault},
		{"plain", stderrs.New("x"), stderr.StepDirectiveDefault},
		{"stop", stderr.NewStopError("c", "d", nil, nil), stderr.StepDirectiveStop},
		{"skip", stderr.NewSkipError("c", "d", nil, nil), stderr.StepDirectiveSkip},
		{"wrapped stop", pkgerrors.Wrap(stderr.NewStopError("c", "d", nil, nil), "ctx"), stderr.StepDirectiveStop},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, stderr.DirectiveOf(tt.err))
		})
	}
}

func TestErrUser_DefaultDirectiveZeroValue(t *testing.T) {
	u := stderr.ErrUser{Err: stderrs.New("x"), Description: "d"}
	assert.Equal(t, stderr.StepDirectiveDefault, u.Directive)
}
