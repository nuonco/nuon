package executeworkflowstep

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestDispatchFailedStatus(t *testing.T) {
	st := DispatchFailedStatus(errors.New("enqueue failed"))
	require.Equal(t, app.StatusError, st.Status)
	require.Equal(t, "execute signal was not created", st.StatusHumanDescription)
	require.Equal(t, "enqueue failed", st.Metadata["reason"])
}
