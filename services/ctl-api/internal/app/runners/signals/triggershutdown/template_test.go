package triggershutdown

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func TestTemplateDecode(t *testing.T) {
	var sd signaldb.SignalData
	require.NoError(t, json.Unmarshal([]byte(`{"type":"trigger_shutdown","data":{"runner_id":"runacme","process_type":"install","process_id":"rpracme"}}`), &sd))
	sig, ok := sd.Signal.(*Signal)
	require.True(t, ok)
	require.Equal(t, "runacme", sig.RunnerID)
	require.Equal(t, "rpracme", sig.ProcessID)

	// legacy template stored before process_id existed
	var legacy signaldb.SignalData
	require.NoError(t, json.Unmarshal([]byte(`{"type":"trigger_shutdown","data":{"runner_id":"runacme","process_type":"install"}}`), &legacy))
	require.Equal(t, "", legacy.Signal.(*Signal).ProcessID)
}
