package telemetryexport

import (
	"context"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/audit"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/settings"
	"github.com/nuonco/nuon/pkg/telemetryexport"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type fakeVendorClient struct {
	nuonrunner.Client
	settings      *models.AppRunnerGroupSettings
	settingsErr   error
	settingsCalls int
	tokenRequests int
}

func (c *fakeVendorClient) GetSettings(context.Context) (*models.AppRunnerGroupSettings, error) {
	c.settingsCalls++
	return c.settings, c.settingsErr
}

func (c *fakeVendorClient) CreateInstallTelemetryAccessToken(context.Context, string, string) (*models.ServiceCreateInstallTelemetryAccessTokenResponse, error) {
	c.tokenRequests++
	return nil, errors.New("unexpected token request")
}

func TestVendorLifecycleRemovesStaleTokenWithoutFetchingSettings(t *testing.T) {
	tests := map[string]*settings.Settings{
		"local runner": {
			Cfg:                    &runnerconfig.Config{IsNuonctl: true},
			Metadata:               map[string]string{"install.id": "inst-test"},
			VendorTelemetryEnabled: true,
			TelemetryRelayEndpoint: "https://relay.example.com",
		},
		"missing install": {
			Cfg:                    &runnerconfig.Config{},
			VendorTelemetryEnabled: true,
			TelemetryRelayEndpoint: "https://relay.example.com",
		},
		"initially disabled install": {
			Cfg:      &runnerconfig.Config{},
			Metadata: map[string]string{"install.id": "inst-test"},
		},
	}
	for name, runnerSettings := range tests {
		t.Run(name, func(t *testing.T) {
			config := telemetryexport.DefaultConfig(t.TempDir())
			if err := os.MkdirAll(config.TokenDirectory, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(config.TokenDirectory, "access-token"), []byte("stale"), 0o600); err != nil {
				t.Fatal(err)
			}
			client := &fakeVendorClient{}
			lifecycle := fxtest.NewLifecycle(t)
			newVendor(VendorParams{Lifecycle: lifecycle, Settings: runnerSettings, Logger: zap.NewNop(), APIClient: client, Audit: &audit.Writer{}}, config)
			lifecycle.RequireStart().RequireStop()

			if _, err := os.Stat(config.TokenDirectory); !os.IsNotExist(err) {
				t.Fatalf("stale relay credentials were not removed: %v", err)
			}
			if client.settingsCalls != 0 || client.tokenRequests != 0 {
				t.Fatalf("ineligible or disabled runner contacted the API: settings=%d tokens=%d", client.settingsCalls, client.tokenRequests)
			}
		})
	}
}

func TestFetchVendorSettingsMapsRunnerSettings(t *testing.T) {
	attributes := map[string]string{"nuon.install.name": "acme"}
	client := &fakeVendorClient{settings: &models.AppRunnerGroupSettings{
		VendorTelemetryEnabled:            true,
		TelemetryRelayEndpoint:            "https://relay.example.com",
		VendorTelemetryResourceAttributes: attributes,
	}}
	got, err := fetchVendorSettings(context.Background(), client)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Enabled || got.RelayEndpoint != "https://relay.example.com" || !maps.Equal(got.ResourceAttributes, attributes) {
		t.Fatalf("runner settings were not mapped: %#v", got)
	}

	client.settings = nil
	if _, err := fetchVendorSettings(context.Background(), client); err == nil {
		t.Fatal("empty runner settings response was accepted")
	}

	client.settingsErr = errors.New("unavailable")
	if _, err := fetchVendorSettings(context.Background(), client); !errors.Is(err, client.settingsErr) {
		t.Fatalf("settings error was not returned: %v", err)
	}
}

type recordingAuditWriter struct {
	events []audit.Event
	err    error
}

func (w *recordingAuditWriter) WriteAsync(event audit.Event) error {
	w.events = append(w.events, event)
	return w.err
}

func TestVendorTelemetryAuditRecordsAppliedSetting(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	writer := &recordingAuditWriter{}
	report := vendorTelemetryAudit(writer, zap.New(core))

	report(true)
	report(false)
	if len(writer.events) != 2 {
		t.Fatalf("expected two audit events, got %d", len(writer.events))
	}
	for i, enabled := range []string{"true", "false"} {
		event := writer.events[i]
		if event.Name != "install_telemetry_updated" || event.Outcome != audit.OutcomeSucceeded || event.Attributes["telemetry.enabled"] != enabled {
			t.Fatalf("unexpected audit event %d: %#v", i, event)
		}
	}

	writer.err = audit.ErrUnavailable
	report(true)
	if logs.Len() != 0 {
		t.Fatal("unavailable audit route was logged as a failure")
	}
	writer.err = errors.New("queue full")
	report(false)
	if logs.Len() != 1 {
		t.Fatalf("audit enqueue failure was not logged: %d", logs.Len())
	}
}
