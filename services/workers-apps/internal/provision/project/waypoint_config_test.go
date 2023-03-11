package project

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getProjectWaypointConfig(t *testing.T) {
	tests := map[string]struct {
		appID       string
		errExpected error
		assertFn    func(*testing.T, []byte)
	}{
		"happy path": {
			appID:       "abc-app-id",
			errExpected: nil,
			assertFn: func(t *testing.T, byts []byte) {
				var vals map[string]string
				err := json.Unmarshal(byts, &vals)
				assert.NoError(t, err)

				assert.Equal(t, "abc-app-id", vals["project"])
			},
		},
		"empty app id": {
			appID:       "",
			errExpected: fmt.Errorf(""),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			byts, err := getProjectWaypointConfig(test.appID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, byts)
		})
	}
}
