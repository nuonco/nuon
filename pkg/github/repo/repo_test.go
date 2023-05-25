package github

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseRepo(t *testing.T) {
	tests := map[string]struct {
		input         string
		nameExpected  string
		ownerExpected string
		errExpected   error
	}{
		"happy path": {
			input:         "powertoolsdev/api",
			nameExpected:  "api",
			ownerExpected: "powertoolsdev",
		},
		"dot github": {
			input:         "powertoolsdev/.github",
			ownerExpected: "powertoolsdev",
			nameExpected:  ".github",
		},
		"invalid just a name or owner": {
			input:       "powertoolsdev",
			errExpected: fmt.Errorf("invalid github repo"),
		},
		"invalid too many slashes": {
			input:       "powertoolsdev/.github/tree",
			errExpected: fmt.Errorf("invalid github repo"),
		},
		"invalid starts with https://": {
			input:       "https://github.com/powertoolsdev/.github/tree",
			errExpected: fmt.Errorf("invalid github repo"),
		},
		"invalid starts with git@": {
			input:       "git@github.com:powertoolsdev/.github.git",
			errExpected: fmt.Errorf("invalid github repo"),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			owner, name, err := parseRepo(test.input)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.ownerExpected, owner)
			assert.Equal(t, test.nameExpected, name)
		})
	}
}
