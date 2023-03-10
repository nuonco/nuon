package ecr

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const fakeEcrImageURL = "766121324316.dkr.ecr.us-west-2.amazonaws.com/httpbin:latest"

func Test_parseImageUrl(t *testing.T) {
	tests := map[string]struct {
		input       string
		expected    string
		errExpected error
	}{
		"happy path": {
			input:    fakeEcrImageURL,
			expected: "766121324316",
		},
		"invalid": {
			input:       uuid.NewString(),
			errExpected: fmt.Errorf("invalid ecr image url"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := parseImageURL(test.input)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expected, resp)
		})
	}
}
