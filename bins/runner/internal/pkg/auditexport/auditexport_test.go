package auditexport

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestReconcilePreservesLastKnownGoodConfiguration(t *testing.T) {
	valid := "exporters:\n  otlphttp:\n    endpoint: https://otlp.example.com\n"
	invalid := "exporters:\n  otlphttp:\n    endpoint: http://otlp.example.com\n"
	replacements := 0
	stops := 0
	s := &Supervisor{
		installID: "inst-test",
		logger:    zap.NewNop(),
		active:    valid,
		replaceChildFn: func(secretConfig) error {
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
	valid := "exporters:\n  otlphttp:\n    endpoint: https://otlp.example.com\n"
	replacements := 0
	s := &Supervisor{
		installID: "inst-test",
		logger:    zap.NewNop(),
		replaceChildFn: func(secretConfig) error {
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
	current := "exporters:\n  otlphttp:\n    endpoint: https://current.example.com\n"
	updated := "exporters:\n  otlphttp:\n    endpoint: https://updated.example.com\n"
	s := &Supervisor{
		installID: "inst-test",
		logger:    zap.NewNop(),
		active:    current,
		backoff:   time.Second,
		replaceChildFn: func(secretConfig) error {
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
	valid := "exporters:\n  otlphttp:\n    endpoint: https://otlp.example.com\n"
	stops := 0
	s := &Supervisor{
		installID:   "inst-test",
		logger:      zap.NewNop(),
		active:      valid,
		enabled:     true,
		reported:    true,
		stopChildFn: func() { stops++ },
	}

	s.reconcile(configUpdate{state: configUnavailable})
	if stops != 1 || s.active != "" || s.enabled {
		t.Fatal("unavailable secret did not disable the collector")
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
		replaceChildFn: func(secretConfig) error {
			replacements++
			return nil
		},
		stopChildFn: func() {},
	}

	s.reconcile(configUpdate{state: configAvailable})
	if replacements != 0 || s.active != "" || s.enabled {
		t.Fatal("empty secret started the audit export collector")
	}
}

func TestReconcileLogsEnabledBackendAndDisabledTransition(t *testing.T) {
	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	valid := "exporters:\n  otlphttp:\n    endpoint: https://otlp.example.com/tenant\n    headers:\n      Authorization: secret-value\n"
	s := &Supervisor{
		installID: "inst-test",
		logger:    logger,
		replaceChildFn: func(secretConfig) error {
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
	if entries[0].Message != "runner audit export enabled" || entries[0].ContextMap()["audit_export.backend"] != "otlp.example.com" {
		t.Fatalf("unexpected enabled log: %#v", entries[0])
	}
	if entries[1].Message != "runner audit export disabled" || entries[1].ContextMap()["audit_export.reason"] != "secret not found" {
		t.Fatalf("unexpected disabled log: %#v", entries[1])
	}
	for _, entry := range entries {
		if entry.ContextMap()["audit_export.backend"] == "secret-value" {
			t.Fatal("audit export log exposed a credential")
		}
	}
}
