package actionerrors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestPreparationFailedError(t *testing.T) {
	err := &PreparationFailedError{Detail: "stack outputs must have either AWS, Azure, or GCP outputs"}

	data, freezeErr := compositeerrors.New(err)
	require.NoError(t, freezeErr)
	assert.Equal(t, PreparationFailedErrorType, data.Type)
	assert.Equal(t, compositeerrors.SeverityError, data.Severity)
	assert.Equal(t, "Unable to prepare action run", data.Message)
	require.Len(t, data.Sections, 3)
	assert.Equal(t, compositeerrors.SectionCode, data.Sections[1].Kind)
	assert.Equal(t, err.Detail, data.Sections[1].Body)
}
