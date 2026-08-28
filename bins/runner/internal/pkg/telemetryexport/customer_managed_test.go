package telemetryexport

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCustomerManagedCollectorConfig(t *testing.T) {
	dir := t.TempDir()
	raw, err := customerManagedCollectorConfig(dir)
	require.NoError(t, err)

	var config map[string]any
	require.NoError(t, yaml.Unmarshal(raw, &config))
	receivers := config["receivers"].(map[string]any)
	otlp := receivers["otlp"].(map[string]any)
	protocols := otlp["protocols"].(map[string]any)
	http := protocols["http"].(map[string]any)
	require.Equal(t, "127.0.0.1:4318", http["endpoint"])

	exporters := config["exporters"].(map[string]any)
	file := exporters["file/logs"].(map[string]any)
	require.Equal(t, filepath.Join(dir, "otel.jsonl"), file["path"])
	require.Equal(t, "json", file["format"])
	rotation := file["rotation"].(map[string]any)
	require.EqualValues(t, 10, rotation["max_megabytes"])
	require.EqualValues(t, 10, rotation["max_backups"])

	service := config["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)
	logs := pipelines["logs"].(map[string]any)
	require.Equal(t, []any{"otlp"}, logs["receivers"])
	require.Equal(t, []any{"file/logs"}, logs["exporters"])
}
