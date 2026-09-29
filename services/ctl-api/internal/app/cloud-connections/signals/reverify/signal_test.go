package reverify

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	for name, tc := range map[string]struct {
		id    string
		valid bool
	}{
		"missing connection": {},
		"connection":         {id: "cc_acme", valid: true},
	} {
		t.Run(name, func(t *testing.T) {
			err := (&Signal{CloudConnectionID: tc.id}).Validate(nil)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, "cloud_connection_id is required")
			}
		})
	}
}
