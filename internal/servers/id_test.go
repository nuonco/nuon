package servers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ensureShortID(t *testing.T) {
	tests := map[string]struct {
		input       string
		output      string
		errExpected error
	}{
		"shortid": {
			input:  "0ujvxnyuwuzqt251ozv0hovma9",
			output: "0ujvxnyuwuzqt251ozv0hovma9",
		},
		"uuid": {
			input:  "37cebef0-e72e-4875-8cbc-4dc3603f3e91",
			output: "0ujvxnyuwuzqt251ozv0hovma9",
		},
		"error": {
			input:       "bad-id",
			errExpected: fmt.Errorf("neither a shortID or UUID"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			val, err := ensureShortID(test.input)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.output, val)
		})
	}
}

type ensureShortIDTest struct {
	IDField    string `shortid:"ensure"`
	NonIDField string
}

func TestEnsureShortID(t *testing.T) {
	tests := map[string]struct {
		input       interface{}
		output      interface{}
		errExpected error
	}{
		"happy path - no changes": {
			input: &ensureShortIDTest{
				IDField:    "0ujvxnyuwuzqt251ozv0hovma9",
				NonIDField: "non-id",
			},
			output: &ensureShortIDTest{
				IDField:    "0ujvxnyuwuzqt251ozv0hovma9",
				NonIDField: "non-id",
			},
		},
		"happy path": {
			input: &ensureShortIDTest{
				IDField:    "37cebef0-e72e-4875-8cbc-4dc3603f3e91",
				NonIDField: "non-id",
			},
			output: &ensureShortIDTest{
				IDField:    "0ujvxnyuwuzqt251ozv0hovma9",
				NonIDField: "non-id",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := EnsureShortID(test.input)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.output, test.input)
		})
	}
}
