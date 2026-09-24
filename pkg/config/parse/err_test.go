package parse

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseErrIncludesAndUnwrapsCause(t *testing.T) {
	cause := errors.New("source values/chart.yaml does not exist")
	err := ParseErr{
		Filename:    "components/chart.toml",
		Description: "error parsing config",
		Err:         cause,
	}

	require.EqualError(t, err, "components/chart.toml: error parsing config: source values/chart.yaml does not exist")
	require.ErrorIs(t, err, cause)
}
