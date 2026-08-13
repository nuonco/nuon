package audit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"github.com/nuonco/nuon/pkg/runner/settings"
	"github.com/nuonco/nuon/pkg/runner/version"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type testExporter struct {
	mu       sync.Mutex
	records  []sdklog.Record
	err      error
	wait     bool
	shutdown bool
}

type testMetrics struct {
	mu          sync.Mutex
	increments  [][]string
	timingCount int
}

func (m *testMetrics) Incr(name string, tags []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.increments = append(m.increments, append([]string{name}, tags...))
}

func (m *testMetrics) Timing(string, time.Duration, []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timingCount++
}

func (e *testExporter) Export(ctx context.Context, records []sdklog.Record) error {
	if e.wait {
		<-ctx.Done()
		return ctx.Err()
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, record := range records {
		e.records = append(e.records, record.Clone())
	}
	return e.err
}

func (e *testExporter) Shutdown(context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.shutdown = true
	return nil
}

func attrs(record sdklog.Record) map[string]string {
	result := make(map[string]string)
	record.WalkAttributes(func(attr otellog.KeyValue) bool {
		if attr.Value.Kind() == otellog.KindString {
			result[attr.Key] = attr.Value.AsString()
		}
		return true
	})
	return result
}

func TestProcessIdentityUsesRegisteredProcessAndConfiguredImage(t *testing.T) {
	previousVersion := version.Version
	version.Version = "v1.2.3"
	t.Cleanup(func() { version.Version = previousVersion })

	identity := processIdentity("proc-123", "install", &settings.Settings{
		Metadata:          map[string]string{"runner.id": "run-123", "runner.type": "install", "runner.platform": "aws", "install.id": "inst-123", "org.id": "org-123"},
		ContainerImageURL: "registry.example.com/runner",
		ContainerImageTag: "stable",
	})
	want := map[string]string{
		"runner_process.id":           "proc-123",
		"runner_process.type":         "install",
		"runner_process.version":      "v1.2.3",
		"runner.id":                   "run-123",
		"runner.type":                 "install",
		"runner.platform":             "aws",
		"install.id":                  "inst-123",
		"org.id":                      "org-123",
		"runner.image.configured_url": "registry.example.com/runner",
		"runner.image.configured_tag": "stable",
		"runner.image.tag_identity":   "configured",
	}
	for key, value := range want {
		if identity[key] != value {
			t.Errorf("%s = %q, want %q", key, identity[key], value)
		}
	}
	if _, ok := identity["runner.image.digest"]; ok {
		t.Fatal("process identity included an unverified image digest")
	}
}

func TestWriteSyncWaitsForAcknowledgementAndReturnsErrors(t *testing.T) {
	syncExporter := new(testExporter)
	w := newWriter(map[string]string{"runner_process.id": "proc-123"}, nil, new(testExporter), syncExporter)
	w.enabled = true

	event := Event{Name: "test", Message: "test event", Outcome: OutcomeSucceeded}
	if err := w.WriteSync(context.Background(), event); err != nil {
		t.Fatalf("WriteSync() error = %v", err)
	}
	if len(syncExporter.records) != 1 {
		t.Fatalf("exported %d records, want 1", len(syncExporter.records))
	}

	syncExporter.err = errors.New("collector rejected record")
	if err := w.WriteSync(context.Background(), event); err == nil {
		t.Fatal("WriteSync() returned nil for collector error")
	}

	syncExporter.err = nil
	syncExporter.wait = true
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := w.WriteSync(ctx, event); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WriteSync() error = %v, want deadline exceeded", err)
	}
}

func TestWriteSyncEnforcesConfiguredTimeout(t *testing.T) {
	syncExporter := &testExporter{wait: true}
	w := newWriter(nil, nil, new(testExporter), syncExporter)
	w.enabled = true
	w.syncTimeout = 10 * time.Millisecond

	err := w.WriteSync(context.Background(), Event{Name: "test", Message: "test event", Outcome: OutcomeSucceeded})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WriteSync() error = %v, want configured deadline exceeded", err)
	}
}

func TestWriteSyncPropagatesCallerCancellation(t *testing.T) {
	syncExporter := &testExporter{wait: true}
	w := newWriter(nil, nil, new(testExporter), syncExporter)
	w.enabled = true

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := w.WriteSync(ctx, Event{Name: "test", Message: "test event", Outcome: OutcomeSucceeded})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WriteSync() error = %v, want caller cancellation", err)
	}
}

func TestWriteAsyncSendsToCollector(t *testing.T) {
	asyncExporter := new(testExporter)
	syncExporter := new(testExporter)
	w := newWriter(nil, nil, asyncExporter, syncExporter)
	w.enabled = true
	before := time.Now()
	if err := w.WriteAsync(Event{Name: "test", Message: "queued", Outcome: OutcomeStarted}); err != nil {
		t.Fatalf("WriteAsync() error = %v", err)
	}
	if len(asyncExporter.records) != 1 {
		t.Fatalf("asynchronous collector route received %d records, want 1", len(asyncExporter.records))
	}
	if len(syncExporter.records) != 0 {
		t.Fatalf("synchronous collector route received %d records, want 0", len(syncExporter.records))
	}
	if timestamp := asyncExporter.records[0].Timestamp(); timestamp.Before(before) || timestamp.After(time.Now()) {
		t.Fatalf("collector record timestamp = %s, want write time", timestamp)
	}
}

func TestOTLPWriteSyncWaitsForCollectorResponse(t *testing.T) {
	var reject atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tenant/v1/logs" {
			t.Errorf("request path = %q, want /tenant/v1/logs", r.URL.Path)
		}
		time.Sleep(40 * time.Millisecond)
		if reject.Load() {
			http.Error(w, "rejected", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	exp, err := newOTLPExporter(context.Background(), server.URL+"/tenant")
	if err != nil {
		t.Fatalf("newOTLPExporter() error = %v", err)
	}
	w := newWriter(nil, nil, new(testExporter), exp)
	w.enabled = true

	started := time.Now()
	if err := w.WriteSync(context.Background(), Event{Name: "test", Message: "acknowledge me", Outcome: OutcomeSucceeded}); err != nil {
		t.Fatalf("WriteSync() error = %v", err)
	}
	if elapsed := time.Since(started); elapsed < 40*time.Millisecond {
		t.Fatalf("WriteSync() returned after %s, before backend acknowledgement", elapsed)
	}

	reject.Store(true)
	if err := w.WriteSync(context.Background(), Event{Name: "test", Message: "reject me", Outcome: OutcomeFailed}); err == nil {
		t.Fatal("WriteSync() returned nil after customer backend rejection")
	}
}

func TestEnableEmitsStartupOnce(t *testing.T) {
	syncExporter := new(testExporter)
	w := newWriter(map[string]string{"runner_process.id": "proc-123"}, nil, new(testExporter), syncExporter)

	w.Enable()
	w.Disable()
	w.Enable()
	if len(syncExporter.records) != 1 || !w.Available() {
		t.Fatalf("startup exports = %d, available = %t; want 1, true", len(syncExporter.records), w.Available())
	}
}

func TestLifecycleEnvelopeAndStoppingAreSynchronousAndDeduplicated(t *testing.T) {
	asyncExporter := new(testExporter)
	syncExporter := new(testExporter)
	w := newWriter(map[string]string{"runner_process.id": "proc-123"}, nil, asyncExporter, syncExporter)
	w.Enable()
	if err := w.ProcessStopping(context.Background(), "host_shutdown", "systemd_logind"); err != nil {
		t.Fatalf("ProcessStopping() error = %v", err)
	}
	if err := w.ProcessStopping(context.Background(), "process_stopping", "fx_lifecycle"); err != nil {
		t.Fatalf("second ProcessStopping() error = %v", err)
	}
	if len(syncExporter.records) != 2 || len(asyncExporter.records) != 0 {
		t.Fatalf("lifecycle exports = sync %d, async %d; want 2, 0", len(syncExporter.records), len(asyncExporter.records))
	}
	started := attrs(syncExporter.records[0])
	stopping := attrs(syncExporter.records[1])
	if started[AttrEvent] != "runner_process_lifecycle" || started[AttrOutcome] != OutcomeStarted || started["runner_process.id"] != "proc-123" || syncExporter.records[0].Timestamp().IsZero() {
		t.Fatalf("invalid startup envelope: %#v", started)
	}
	if stopping[AttrOutcome] != OutcomeStopping || stopping["runner.shutdown.reason"] != "host_shutdown" || stopping["runner.shutdown.source"] != "systemd_logind" {
		t.Fatalf("invalid stopping envelope: %#v", stopping)
	}
}

func TestProcessStoppingRetriesAfterFailedExport(t *testing.T) {
	syncExporter := &testExporter{err: errors.New("collector rejected record")}
	w := newWriter(nil, nil, new(testExporter), syncExporter)
	w.enabled = true

	if err := w.ProcessStopping(context.Background(), "host_shutdown", "systemd_logind"); err == nil {
		t.Fatal("ProcessStopping() returned nil for failed export")
	}
	syncExporter.mu.Lock()
	syncExporter.err = nil
	syncExporter.mu.Unlock()
	if err := w.ProcessStopping(context.Background(), "graceful_shutdown", "fx_lifecycle"); err != nil {
		t.Fatalf("ProcessStopping() retry error = %v", err)
	}
	if err := w.ProcessStopping(context.Background(), "graceful_shutdown", "fx_lifecycle"); err != nil {
		t.Fatalf("ProcessStopping() after success error = %v", err)
	}
	if len(syncExporter.records) != 2 {
		t.Fatalf("stopping export attempts = %d, want 2", len(syncExporter.records))
	}
}

func TestProcessStoppingSerializesConcurrentCalls(t *testing.T) {
	syncExporter := new(testExporter)
	w := newWriter(nil, nil, new(testExporter), syncExporter)
	w.enabled = true

	const callers = 16
	start := make(chan struct{})
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- w.ProcessStopping(context.Background(), "graceful_shutdown", "fx_lifecycle")
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("ProcessStopping() error = %v", err)
		}
	}
	if len(syncExporter.records) != 1 {
		t.Fatalf("successful stopping exports = %d, want 1", len(syncExporter.records))
	}
}

func TestWritesRecordDeliveryMetrics(t *testing.T) {
	metrics := new(testMetrics)
	asyncExporter := new(testExporter)
	syncExporter := new(testExporter)
	w := newWriter(nil, metrics, asyncExporter, syncExporter)
	w.enabled = true

	event := Event{Name: "test", Message: "test event", Outcome: OutcomeSucceeded}
	if err := w.WriteAsync(event); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteSync(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	syncExporter.err = errors.New("collector rejected record")
	if err := w.WriteSync(context.Background(), event); err == nil {
		t.Fatal("expected synchronous failure")
	}
	syncExporter.err = nil
	syncExporter.wait = true
	w.syncTimeout = 10 * time.Millisecond
	if err := w.WriteSync(context.Background(), event); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := w.WriteSync(ctx, event); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
	w.Disable()
	if err := w.WriteAsync(event); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("unavailable error = %v", err)
	}

	want := [][]string{
		{"runner.audit_export.write", "route:async", "status:success"},
		{"runner.audit_export.write", "route:sync", "status:success"},
		{"runner.audit_export.write", "route:sync", "status:failure"},
		{"runner.audit_export.write", "route:sync", "status:timeout"},
		{"runner.audit_export.write", "route:sync", "status:canceled"},
		{"runner.audit_export.write", "route:async", "status:unavailable"},
	}
	if len(metrics.increments) != len(want) {
		t.Fatalf("audit export metric count = %d, want %d: %v", len(metrics.increments), len(want), metrics.increments)
	}
	for idx := range want {
		if strings.Join(metrics.increments[idx], ",") != strings.Join(want[idx], ",") {
			t.Errorf("audit export metric %d = %v, want %v", idx, metrics.increments[idx], want[idx])
		}
	}
	if metrics.timingCount != len(want) {
		t.Fatalf("audit export timing count = %d, want %d", metrics.timingCount, len(want))
	}
}

func TestJobEventPreservesExistingEnvelope(t *testing.T) {
	event, ok := JobEvent(&models.AppRunnerJob{
		Group:           models.AppRunnerJobGroupDeploy,
		CreatedByID:     "acct-123",
		OrgID:           "org-123",
		RunnerProcessID: "proc-123",
		Metadata:        map[string]string{"install_id": "inst-123"},
	}, "job execution started", OutcomeStarted, nil)
	if !ok {
		t.Fatal("deploy job was not auditable")
	}
	if event.Name != "install_deploy" || event.Attributes["user.id"] != "acct-123" || event.Attributes["install.id"] != "inst-123" || event.Attributes["runner_process.id"] != "proc-123" {
		t.Fatalf("unexpected job audit event: %#v", event)
	}
	if _, ok := JobEvent(&models.AppRunnerJob{Group: models.AppRunnerJobGroupBuild}, "ignored", OutcomeStarted, nil); ok {
		t.Fatal("build job was auditable")
	}
}
