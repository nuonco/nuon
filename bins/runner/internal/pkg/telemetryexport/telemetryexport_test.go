package telemetryexport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type recordingAuditConfigurator struct {
	enabled    int
	disabled   int
	stopping   int
	enableErr  error
	onStopping func()
}

func (c *recordingAuditConfigurator) Enable() error {
	c.enabled++
	return c.enableErr
}

func (c *recordingAuditConfigurator) Disable() {
	c.disabled++
}

func (c *recordingAuditConfigurator) ProcessStopping(context.Context, string, string) error {
	c.stopping++
	if c.onStopping != nil {
		c.onStopping()
	}
	return nil
}

func TestWaitForCollector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := waitForCollector(&childProcess{done: make(chan struct{})}, server.URL); err != nil {
		t.Fatalf("waitForCollector() error = %v", err)
	}
}

func TestWaitForCollectorDetectsExitedProcess(t *testing.T) {
	done := make(chan struct{})
	close(done)
	if err := waitForCollector(&childProcess{done: done}, "http://127.0.0.1:0"); err == nil {
		t.Fatal("waitForCollector() returned nil for exited process")
	}
}

func TestStopWritesAuditBeforeStoppingCollector(t *testing.T) {
	var order []string
	done := make(chan struct{})
	close(done)
	s := &Supervisor{
		logger: zap.NewNop(),
		audit: &recordingAuditConfigurator{onStopping: func() {
			order = append(order, "audit")
		}},
		cancel: func() {
			order = append(order, "collector")
		},
		done: done,
	}
	if err := s.stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(order) != 2 || order[0] != "audit" || order[1] != "collector" {
		t.Fatalf("shutdown order = %q, want audit before collector", order)
	}
}

func TestReconcilePreservesLastKnownGoodConfiguration(t *testing.T) {
	valid := auditEnabledConfig("https://otlp.example.com")
	invalid := "version: v2\n"
	replacements := 0
	stops := 0
	s := &Supervisor{
		installID: "inst-test",
		logger:    zap.NewNop(),
		active:    valid,
		replaceChildFn: func(config) error {
			replacements++
			return nil
		},
		stopChildFn: func() { stops++ },
	}

	s.reconcile(configUpdate{state: configAvailable, value: invalid})
	if replacements != 0 || stops != 0 || s.active != valid {
		t.Fatal("invalid update changed the active collector")
	}

	s.reconcile(configUpdate{state: configLookupFailed, err: errors.New("temporary failure")})
	if replacements != 0 || stops != 0 || s.active != valid {
		t.Fatal("transient lookup failure changed the active collector")
	}

	s.reconcile(configUpdate{state: configNotFound})
	if replacements != 0 || stops != 1 || s.active != "" {
		t.Fatal("missing secret did not disable the collector")
	}
}

func TestReconcileAppliesChangedValidConfigurationOnce(t *testing.T) {
	valid := auditEnabledConfig("https://otlp.example.com")
	replacements := 0
	s := &Supervisor{
		installID: "inst-test",
		logger:    zap.NewNop(),
		replaceChildFn: func(config) error {
			replacements++
			return nil
		},
		stopChildFn: func() {},
	}

	s.reconcile(configUpdate{state: configAvailable, value: valid})
	s.reconcile(configUpdate{state: configAvailable, value: valid})
	if replacements != 1 || s.active != valid {
		t.Fatal("valid configuration was not applied exactly once")
	}
}

func TestReconcileEnablesAuditOnlyAfterCollectorReplacement(t *testing.T) {
	first := auditEnabledConfig("https://first.example.com")
	second := auditEnabledConfig("https://second.example.com")
	auditWriter := new(recordingAuditConfigurator)
	replacements := 0
	s := &Supervisor{
		installID: "inst-test",
		logger:    zap.NewNop(),
		audit:     auditWriter,
		replaceChildFn: func(config) error {
			replacements++
			if replacements == 1 {
				return errors.New("collector unavailable")
			}
			return nil
		},
		stopChildFn: func() {},
	}

	s.reconcile(configUpdate{state: configAvailable, value: first})
	if auditWriter.enabled != 0 || s.active != "" {
		t.Fatal("failed collector replacement enabled audit delivery")
	}
	s.reconcile(configUpdate{state: configAvailable, value: first})
	s.reconcile(configUpdate{state: configAvailable, value: second})
	if auditWriter.enabled != 2 || s.active != second {
		t.Fatalf("successful collector replacements enabled audit %d times with active config %q", auditWriter.enabled, s.active)
	}
	s.reconcile(configUpdate{state: configNotFound})
	if auditWriter.disabled != 1 {
		t.Fatalf("disabled audit exporter %d times, want 1", auditWriter.disabled)
	}
}

func TestReconcileSchedulesRestartWhenUpdateAndRollbackFail(t *testing.T) {
	current := auditEnabledConfig("https://current.example.com")
	updated := auditEnabledConfig("https://updated.example.com")
	s := &Supervisor{
		installID: "inst-test",
		logger:    zap.NewNop(),
		active:    current,
		backoff:   time.Second,
		replaceChildFn: func(config) error {
			return errors.New("start failed")
		},
		stopChildFn: func() {},
	}

	s.reconcile(configUpdate{state: configAvailable, value: updated})
	if s.nextStart.IsZero() || s.active != current {
		t.Fatal("failed rollback did not schedule recovery of the last-known-good configuration")
	}
}

func TestReconcileFailedUpdateKeepsAuditOnLastKnownGoodCollector(t *testing.T) {
	current := auditEnabledConfig("https://current.example.com")
	updated := auditEnabledConfig("https://updated.example.com")
	auditWriter := &recordingAuditConfigurator{enabled: 1}
	var endpoints []string
	s := &Supervisor{
		installID:        "inst-test",
		logger:           zap.NewNop(),
		active:           current,
		collectorEnabled: true,
		audit:            auditWriter,
		replaceChildFn: func(cfg config) error {
			endpoints = append(endpoints, cfg.OTLPHTTP.Endpoint)
			if cfg.OTLPHTTP.Endpoint == "https://updated.example.com" {
				return errors.New("start failed")
			}
			return nil
		},
		stopChildFn: func() {},
	}

	s.reconcile(configUpdate{state: configAvailable, value: updated})
	if s.active != current || auditWriter.enabled != 1 || len(endpoints) != 2 || endpoints[0] != "https://updated.example.com" || endpoints[1] != "https://current.example.com" {
		t.Fatalf("failed update changed audit routing: active=%q enabled=%d endpoints=%q", s.active, auditWriter.enabled, endpoints)
	}
}

func TestReconcileDisablesUnavailableSecret(t *testing.T) {
	valid := auditEnabledConfig("https://otlp.example.com")
	stops := 0
	auditWriter := new(recordingAuditConfigurator)
	s := &Supervisor{
		installID:        "inst-test",
		logger:           zap.NewNop(),
		active:           valid,
		collectorEnabled: true,
		audit:            auditWriter,
		reported:         true,
		stopChildFn:      func() { stops++ },
	}

	s.reconcile(configUpdate{state: configUnavailable})
	if stops != 1 || s.active != "" || s.collectorEnabled || auditWriter.disabled != 1 {
		t.Fatal("unavailable secret did not disable the collector")
	}
}

func TestReconcileKeepsCollectorWithoutAuditPipeline(t *testing.T) {
	active := auditEnabledConfig("https://otlp.example.com")
	disabledAudit := "version: v1\ntelemetry:\n  logs:\n    audit:\n      enabled: false\n"
	replacements := 0
	stops := 0
	auditWriter := new(recordingAuditConfigurator)
	s := &Supervisor{
		installID:        "inst-test",
		logger:           zap.NewNop(),
		active:           active,
		nextStart:        time.Now().Add(time.Minute),
		collectorEnabled: true,
		audit:            auditWriter,
		reported:         true,
		replaceChildFn: func(config) error {
			replacements++
			return nil
		},
		stopChildFn: func() { stops++ },
	}

	s.reconcile(configUpdate{state: configAvailable, value: disabledAudit})
	if replacements != 1 || stops != 0 || s.active != disabledAudit || !s.nextStart.IsZero() || !s.collectorEnabled || auditWriter.disabled != 1 {
		t.Fatal("disabled audit logs did not leave the collector running without the audit pipeline")
	}
}

func TestRunDoesNothingInLocalDevelopment(t *testing.T) {
	s := &Supervisor{
		installID: "inst-test",
		platform:  "aws",
		local:     true,
		done:      make(chan struct{}),
	}

	s.run(context.Background())
	select {
	case <-s.done:
	default:
		t.Fatal("local audit export supervisor did not exit")
	}
}

func TestEmptySecretDoesNotStartCollector(t *testing.T) {
	replacements := 0
	s := &Supervisor{
		installID: "inst-test",
		logger:    zap.NewNop(),
		replaceChildFn: func(config) error {
			replacements++
			return nil
		},
		stopChildFn: func() {},
	}

	s.reconcile(configUpdate{state: configAvailable})
	if replacements != 0 || s.active != "" || s.collectorEnabled {
		t.Fatal("empty secret started the audit export collector")
	}
}

func TestReconcileLogsEnabledBackendAndDisabledTransition(t *testing.T) {
	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	valid := auditEnabledConfig("https://otlp.example.com/tenant") + "    headers:\n      Authorization: secret-value\n"
	s := &Supervisor{
		installID: "inst-test",
		logger:    logger,
		replaceChildFn: func(config) error {
			return nil
		},
		stopChildFn: func() {},
	}

	s.reconcile(configUpdate{state: configAvailable, value: valid})
	s.reconcile(configUpdate{state: configNotFound})
	s.reconcile(configUpdate{state: configNotFound})

	entries := observed.All()
	if len(entries) != 2 {
		t.Fatalf("expected enabled and disabled logs, got %d", len(entries))
	}
	if entries[0].Message != "runner telemetry export collector enabled" || entries[0].ContextMap()["audit_export.backend"] != "otlp.example.com" {
		t.Fatalf("unexpected enabled log: %#v", entries[0])
	}
	if entries[1].Message != "runner telemetry export collector disabled" || entries[1].ContextMap()["telemetry_export.reason"] != "secret not found" {
		t.Fatalf("unexpected disabled log: %#v", entries[1])
	}
	for _, entry := range entries {
		if entry.ContextMap()["audit_export.backend"] == "secret-value" {
			t.Fatal("audit export log exposed a credential")
		}
	}
}

func TestReconcileLogsUnverifiedStartupDeliveryWithoutDisablingCollector(t *testing.T) {
	core, observed := observer.New(zap.InfoLevel)
	valid := auditEnabledConfig("https://otlp.example.com/tenant") + "    headers:\n      Authorization: secret-value\n"
	auditWriter := &recordingAuditConfigurator{enableErr: errors.New("backend echoed secret-value")}
	s := &Supervisor{
		installID: "inst-test",
		logger:    zap.New(core),
		audit:     auditWriter,
		replaceChildFn: func(config) error {
			return nil
		},
		stopChildFn: func() {},
	}

	s.reconcile(configUpdate{state: configAvailable, value: valid})
	if s.active != valid || auditWriter.enabled != 1 {
		t.Fatal("startup delivery failure disabled the active audit configuration")
	}
	entries := observed.FilterMessage("customer audit collector is active but the startup event was not acknowledged; the backend may be unavailable or rejecting credentials").All()
	if len(entries) != 1 {
		t.Fatalf("startup delivery failure logs = %d, want 1", len(entries))
	}
	fields := entries[0].ContextMap()
	if fields["audit_export.backend"] != "otlp.example.com" || fields["audit_export.delivery_verified"] != false || fields["audit_export.failure_type"] != "failure" {
		t.Fatalf("unexpected startup delivery failure log: %#v", fields)
	}
	if strings.Contains(fmt.Sprint(fields), "secret-value") {
		t.Fatal("startup delivery failure log exposed a credential")
	}
}

func TestReconcileLogsCollectorWithoutAuditPipeline(t *testing.T) {
	core, observed := observer.New(zap.InfoLevel)
	s := &Supervisor{
		installID: "inst-test",
		logger:    zap.New(core),
		replaceChildFn: func(config) error {
			return nil
		},
		stopChildFn: func() {},
	}

	s.reconcile(configUpdate{state: configAvailable, value: "version: v1\n"})

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("expected one collector enabled log, got %d", len(entries))
	}
	entry := entries[0]
	fields := entry.ContextMap()
	if entry.Message != "runner telemetry export collector enabled" || fields["telemetry_export.enabled"] != true || fields["audit_export.enabled"] != false {
		t.Fatalf("unexpected collector enabled log: %#v", entry)
	}
	if _, ok := fields["audit_export.backend"]; ok {
		t.Fatal("collector without an audit pipeline logged an audit backend")
	}
}
