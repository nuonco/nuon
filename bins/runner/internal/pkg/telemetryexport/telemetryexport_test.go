package telemetryexport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

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

func TestStopStopsCollector(t *testing.T) {
	stopped := false
	done := make(chan struct{})
	close(done)
	s := &Supervisor{
		logger: zap.NewNop(),
		cancel: func() {
			stopped = true
		},
		done: done,
	}
	if err := s.stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !stopped {
		t.Fatal("collector supervisor was not stopped")
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
	if replacements != 0 || stops != 0 || s.active != valid || s.rejectedConfig != invalid {
		t.Fatal("invalid update changed the active collector")
	}
	s.reconcile(configUpdate{state: configAvailable, value: invalid})
	if replacements != 0 || stops != 0 {
		t.Fatal("unchanged invalid update was retried")
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
	if replacements != 1 || s.active != valid || s.rejectedConfig != "" {
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
	if s.nextStart.IsZero() || s.active != current || s.rejectedConfig != updated || s.restartConfig != current {
		t.Fatal("failed rollback did not schedule recovery of the last-known-good configuration")
	}
}

func TestReconcileSchedulesRestartAfterInitialStartFailure(t *testing.T) {
	value := auditEnabledConfig("https://otlp.example.com")
	attempts := 0
	s := &Supervisor{
		installID: "inst-test",
		logger:    zap.NewNop(),
		backoff:   time.Second,
		replaceChildFn: func(config) error {
			attempts++
			return errors.New("start failed")
		},
		stopChildFn: func() {},
	}

	s.reconcile(configUpdate{state: configAvailable, value: value})
	s.reconcile(configUpdate{state: configAvailable, value: value})
	if attempts != 1 || s.active != "" || s.restartConfig != value || s.nextStart.IsZero() {
		t.Fatalf("initial failure did not schedule restart: active=%q restart=%q next=%s", s.active, s.restartConfig, s.nextStart)
	}
}

func TestRestartActivatesPendingConfiguration(t *testing.T) {
	value := auditEnabledConfig("https://otlp.example.com")
	restarts := 0
	s := &Supervisor{
		installID:     "inst-test",
		logger:        zap.NewNop(),
		restartConfig: value,
		nextStart:     time.Now().Add(-time.Second),
		backoff:       time.Second,
		replaceChildFn: func(config) error {
			restarts++
			return nil
		},
		stopChildFn: func() {},
	}

	s.restartCrashed(context.Background())
	if restarts != 1 || s.active != value || s.restartConfig != "" || !s.nextStart.IsZero() || !s.collectorEnabled {
		t.Fatalf("pending configuration was not activated: restarts=%d active=%q pending=%q next=%s enabled=%t", restarts, s.active, s.restartConfig, s.nextStart, s.collectorEnabled)
	}
}

func TestReconcileFailedUpdateRestoresLastKnownGoodCollector(t *testing.T) {
	current := auditEnabledConfig("https://current.example.com")
	updated := auditEnabledConfig("https://updated.example.com")
	var endpoints []string
	s := &Supervisor{
		installID:        "inst-test",
		logger:           zap.NewNop(),
		active:           current,
		collectorEnabled: true,
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
	s.reconcile(configUpdate{state: configAvailable, value: updated})
	if s.active != current || s.rejectedConfig != updated || len(endpoints) != 2 || endpoints[0] != "https://updated.example.com" || endpoints[1] != "https://current.example.com" {
		t.Fatalf("failed update did not restore the active collector: active=%q endpoints=%q", s.active, endpoints)
	}
}

func TestReconcileRetriesRejectedConfigurationAfterSecretChanges(t *testing.T) {
	current := auditEnabledConfig("https://current.example.com")
	rejected := auditEnabledConfig("https://rejected.example.com")
	corrected := auditEnabledConfig("https://corrected.example.com")
	var endpoints []string
	s := &Supervisor{
		installID: "inst-test",
		logger:    zap.NewNop(),
		active:    current,
		replaceChildFn: func(cfg config) error {
			endpoints = append(endpoints, cfg.OTLPHTTP.Endpoint)
			if cfg.OTLPHTTP.Endpoint == "https://rejected.example.com" {
				return errors.New("start failed")
			}
			return nil
		},
		stopChildFn: func() {},
	}

	s.reconcile(configUpdate{state: configAvailable, value: rejected})
	s.reconcile(configUpdate{state: configAvailable, value: corrected})
	if s.active != corrected || s.rejectedConfig != "" || len(endpoints) != 3 {
		t.Fatalf("changed secret did not replace a rejected configuration: active=%q rejected=%q endpoints=%q", s.active, s.rejectedConfig, endpoints)
	}
}

func TestReconcileDisablesUnavailableSecret(t *testing.T) {
	valid := auditEnabledConfig("https://otlp.example.com")
	stops := 0
	s := &Supervisor{
		installID:        "inst-test",
		logger:           zap.NewNop(),
		active:           valid,
		rejectedConfig:   "rejected",
		restartConfig:    valid,
		nextStart:        time.Now().Add(time.Minute),
		collectorEnabled: true,
		reported:         true,
		stopChildFn:      func() { stops++ },
	}

	s.reconcile(configUpdate{state: configUnavailable})
	if stops != 1 || s.active != "" || s.rejectedConfig != "" || s.restartConfig != "" || !s.nextStart.IsZero() || s.collectorEnabled {
		t.Fatal("unavailable secret did not disable the collector")
	}
}

func TestReconcileKeepsCollectorWithoutAuditPipeline(t *testing.T) {
	active := auditEnabledConfig("https://otlp.example.com")
	disabledAudit := "version: v1\ntelemetry:\n  logs:\n    audit:\n      enabled: false\n"
	replacements := 0
	stops := 0
	s := &Supervisor{
		installID:        "inst-test",
		logger:           zap.NewNop(),
		active:           active,
		restartConfig:    active,
		nextStart:        time.Now().Add(time.Minute),
		collectorEnabled: true,
		reported:         true,
		replaceChildFn: func(config) error {
			replacements++
			return nil
		},
		stopChildFn: func() { stops++ },
	}

	s.reconcile(configUpdate{state: configAvailable, value: disabledAudit})
	if replacements != 1 || stops != 0 || s.active != disabledAudit || s.restartConfig != "" || !s.nextStart.IsZero() || !s.collectorEnabled {
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
		t.Fatal("local telemetry export supervisor did not exit")
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
		t.Fatal("empty secret started the telemetry export collector")
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
