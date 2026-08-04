package telemetryexport

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

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

func TestReconcileDisablesUnavailableSecret(t *testing.T) {
	valid := auditEnabledConfig("https://otlp.example.com")
	stops := 0
	s := &Supervisor{
		installID:          "inst-test",
		logger:             zap.NewNop(),
		active:             valid,
		collectorEnabled:   true,
		auditExportEnabled: true,
		reported:           true,
		stopChildFn:        func() { stops++ },
	}

	s.reconcile(configUpdate{state: configUnavailable})
	if stops != 1 || s.active != "" || s.collectorEnabled || s.auditExportEnabled {
		t.Fatal("unavailable secret did not disable the collector")
	}
}

func TestReconcileKeepsCollectorWithoutAuditPipeline(t *testing.T) {
	active := auditEnabledConfig("https://otlp.example.com")
	disabledAudit := "version: v1\ntelemetry:\n  logs:\n    audit:\n      enabled: false\n"
	replacements := 0
	stops := 0
	s := &Supervisor{
		installID:          "inst-test",
		logger:             zap.NewNop(),
		active:             active,
		nextStart:          time.Now().Add(time.Minute),
		collectorEnabled:   true,
		auditExportEnabled: true,
		reported:           true,
		replaceChildFn: func(config) error {
			replacements++
			return nil
		},
		stopChildFn: func() { stops++ },
	}

	s.reconcile(configUpdate{state: configAvailable, value: disabledAudit})
	if replacements != 1 || stops != 0 || s.active != disabledAudit || !s.nextStart.IsZero() || !s.collectorEnabled || s.auditExportEnabled {
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
	if replacements != 0 || s.active != "" || s.collectorEnabled || s.auditExportEnabled {
		t.Fatal("empty secret started the audit export collector")
	}
}

func TestAuditExportAvailableTracksPipelineAndCollectorProcess(t *testing.T) {
	s := &Supervisor{}
	if s.AuditExportAvailable() {
		t.Fatal("supervisor without a collector is available")
	}

	done := make(chan struct{})
	s.child = &childProcess{done: done}
	if s.AuditExportAvailable() {
		t.Fatal("collector without an audit pipeline reported audit export as available")
	}

	s.auditExportEnabled = true
	if !s.AuditExportAvailable() {
		t.Fatal("running collector is unavailable")
	}

	close(done)
	if s.AuditExportAvailable() {
		t.Fatal("exited collector is available")
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
