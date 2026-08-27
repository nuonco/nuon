package monitor

import (
	"strings"
	"testing"
)

func TestRunnerServiceTemplatesPersistTelemetryExportStorage(t *testing.T) {
	for platform, service := range map[string]string{
		"aws":   runnerServiceAWS,
		"azure": runnerServiceAzure,
		"gcp":   runnerServiceGCP,
	} {
		t.Run(platform, func(t *testing.T) {
			if !strings.Contains(service, telemetryExportStorageMount) {
				t.Fatalf("runner service does not mount persistent telemetry storage: %s", service)
			}
			if !strings.Contains(service, "install -d -m 0700 "+TelemetryExportStorageDirectory) {
				t.Fatalf("runner service does not create protected telemetry storage: %s", service)
			}
			if strings.Contains(service, "find "+TelemetryExportStorageDirectory) {
				t.Fatalf("runner service cleanup removes telemetry storage: %s", service)
			}
		})
	}
}
