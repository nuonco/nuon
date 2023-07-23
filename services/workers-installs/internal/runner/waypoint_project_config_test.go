package runner

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getProjectWaypointConfig(t *testing.T) {
	tests := map[string]struct {
		installID   string
		errExpected error
		assertFn    func(*testing.T, []byte)
	}{
		"hinstally path": {
			installID:   "abc-install-id",
			errExpected: nil,
			assertFn: func(t *testing.T, byts []byte) {
				var vals map[string]string
				err := json.Unmarshal(byts, &vals)
				assert.NoError(t, err)

				assert.Equal(t, "abc-install-id", vals["project"])
			},
		},
		"empty install id": {
			installID:   "",
			errExpected: fmt.Errorf(""),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			byts, err := getProjectWaypointConfig(test.installID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, byts)
		})
	}
}
