package service

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A name the push endpoint would refuse can never be reported, so gating a
// deploy on it would hang until timeout. Reject it at config time instead.
func TestValidateRequiredChecks(t *testing.T) {
	require.NoError(t, validation.ValidateRequiredChecks(nil))
	require.NoError(t, validation.ValidateRequiredChecks([]string{"migrations-applied", "smoke.test_1"}))

	for _, bad := range []string{"", "-leading", "trailing-", "has space", "has/slash"} {
		assert.Error(t, validation.ValidateRequiredChecks([]string{bad}), bad)
	}

	assert.ErrorContains(t, validation.ValidateRequiredChecks([]string{"dup", "dup"}), "duplicate_required_check")
}
