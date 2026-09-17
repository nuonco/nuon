package telemetryexport

import (
	"context"
	"errors"
	"maps"
	"testing"
	"time"

	"go.uber.org/zap"
)

type fakeTokenLifecycle struct {
	enableErr error
	enables   int
	disables  int
	events    *[]string
}

func (f *fakeTokenLifecycle) Enable(context.Context) error {
	f.enables++
	if f.events != nil {
		*f.events = append(*f.events, "token")
	}
	return f.enableErr
}

func (f *fakeTokenLifecycle) Disable() {
	f.disables++
	if f.events != nil {
		*f.events = append(*f.events, "disable-token")
	}
}

func TestVendorSupervisorStartsTokenBeforeCollector(t *testing.T) {
	events := make([]string, 0, 2)
	tokens := &fakeTokenLifecycle{events: &events}
	s := newVendorTestSupervisor(tokens)
	s.replaceChildFn = func(_ context.Context, endpoint string, _ map[string]string) error {
		if endpoint != "https://relay.example.com" {
			t.Fatalf("unexpected endpoint: %q", endpoint)
		}
		events = append(events, "collector")
		return nil
	}

	s.reconcile(context.Background(), vendorSettings{enabled: true, endpoint: "https://relay.example.com"})
	if !s.enabled || s.activeEndpoint != "https://relay.example.com" || tokens.enables != 1 || tokens.disables != 0 || len(events) != 2 || events[0] != "token" || events[1] != "collector" {
		t.Fatalf("vendor collector started without a token: enabled=%t active=%q events=%q", s.enabled, s.activeEndpoint, events)
	}
}

func TestVendorSupervisorEnablesAfterBeingDisabled(t *testing.T) {
	tokens := &fakeTokenLifecycle{}
	replacements := 0
	s := newVendorTestSupervisor(tokens)
	s.replaceChildFn = func(context.Context, string, map[string]string) error {
		replacements++
		return nil
	}

	s.reconcile(context.Background(), vendorSettings{})
	if !s.disabled || tokens.disables != 1 || replacements != 0 {
		t.Fatalf("initially disabled configuration activated collector: disabled=%t disables=%d replacements=%d", s.disabled, tokens.disables, replacements)
	}

	s.reconcile(context.Background(), vendorSettings{enabled: true, endpoint: "https://relay.example.com"})
	if s.disabled || !s.enabled || s.activeEndpoint != "https://relay.example.com" || tokens.enables != 1 || replacements != 1 {
		t.Fatalf("enabled configuration did not activate collector: disabled=%t enabled=%t active=%q enables=%d replacements=%d", s.disabled, s.enabled, s.activeEndpoint, tokens.enables, replacements)
	}
}

func TestVendorSupervisorReplacesCollectorWhenEndpointChanges(t *testing.T) {
	tokens := &fakeTokenLifecycle{}
	var endpoints []string
	s := newVendorTestSupervisor(tokens)
	s.replaceChildFn = func(_ context.Context, endpoint string, _ map[string]string) error {
		endpoints = append(endpoints, endpoint)
		return nil
	}

	s.reconcile(context.Background(), vendorSettings{enabled: true, endpoint: "https://relay-one.example.com"})
	s.reconcile(context.Background(), vendorSettings{enabled: true, endpoint: "https://relay-two.example.com"})
	if !s.enabled || s.activeEndpoint != "https://relay-two.example.com" || len(endpoints) != 2 || endpoints[0] != "https://relay-one.example.com" || endpoints[1] != "https://relay-two.example.com" {
		t.Fatalf("endpoint change was not reconciled: enabled=%t active=%q endpoints=%q", s.enabled, s.activeEndpoint, endpoints)
	}
}

func TestVendorSupervisorDisableStopsCollectorAndToken(t *testing.T) {
	tokens := &fakeTokenLifecycle{}
	stops := 0
	s := newVendorTestSupervisor(tokens)
	s.replaceChildFn = func(context.Context, string, map[string]string) error { return nil }
	s.stopChildFn = func() { stops++ }
	s.reconcile(context.Background(), vendorSettings{enabled: true, endpoint: "https://relay.example.com"})

	s.reconcile(context.Background(), vendorSettings{enabled: false, endpoint: "https://relay.example.com"})
	if !s.disabled || s.enabled || s.activeEndpoint != "" || s.desiredEndpoint != "" || stops != 1 || tokens.disables != 1 {
		t.Fatalf("disabled configuration retained collector state: disabled=%t enabled=%t active=%q desired=%q stops=%d disables=%d", s.disabled, s.enabled, s.activeEndpoint, s.desiredEndpoint, stops, tokens.disables)
	}

	s.reconcile(context.Background(), vendorSettings{enabled: false, endpoint: "https://relay.example.com"})
	if stops != 1 || tokens.disables != 1 {
		t.Fatalf("unchanged disabled configuration repeated cleanup: stops=%d disables=%d", stops, tokens.disables)
	}
}

func TestVendorSupervisorRollsBackFailedEndpointChange(t *testing.T) {
	tokens := &fakeTokenLifecycle{}
	var endpoints []string
	s := newVendorTestSupervisor(tokens)
	s.replaceChildFn = func(_ context.Context, endpoint string, _ map[string]string) error {
		endpoints = append(endpoints, endpoint)
		if endpoint == "https://relay-two.example.com" {
			return errors.New("collector failed")
		}
		return nil
	}
	s.reconcile(context.Background(), vendorSettings{enabled: true, endpoint: "https://relay-one.example.com"})

	s.reconcile(context.Background(), vendorSettings{enabled: true, endpoint: "https://relay-two.example.com"})
	if !s.enabled || s.activeEndpoint != "https://relay-one.example.com" || s.desiredEndpoint != "https://relay-two.example.com" || s.nextStart.IsZero() {
		t.Fatalf("failed replacement did not retain last-known-good endpoint: enabled=%t active=%q desired=%q next=%s", s.enabled, s.activeEndpoint, s.desiredEndpoint, s.nextStart)
	}
	if len(endpoints) != 3 || endpoints[0] != "https://relay-one.example.com" || endpoints[1] != "https://relay-two.example.com" || endpoints[2] != "https://relay-one.example.com" {
		t.Fatalf("unexpected replacement and rollback sequence: %q", endpoints)
	}
	if tokens.disables != 0 {
		t.Fatalf("failed endpoint replacement disabled a token still used by rollback: %d", tokens.disables)
	}
}

func TestVendorSupervisorRejectsInvalidEndpointWithoutStoppingActiveCollector(t *testing.T) {
	tokens := &fakeTokenLifecycle{}
	replacements := 0
	s := newVendorTestSupervisor(tokens)
	s.replaceChildFn = func(context.Context, string, map[string]string) error {
		replacements++
		return nil
	}
	s.reconcile(context.Background(), vendorSettings{enabled: true, endpoint: "https://relay.example.com"})
	s.desiredEndpoint = "https://relay-pending.example.com"
	s.scheduleRestart()

	s.reconcile(context.Background(), vendorSettings{enabled: true, endpoint: "http://insecure.example.com"})
	if !s.enabled || s.activeEndpoint != "https://relay.example.com" || s.desiredEndpoint != "https://relay.example.com" || !s.nextStart.IsZero() || replacements != 1 || tokens.disables != 0 {
		t.Fatalf("invalid endpoint replaced active configuration: enabled=%t active=%q desired=%q next=%s replacements=%d disables=%d", s.enabled, s.activeEndpoint, s.desiredEndpoint, s.nextStart, replacements, tokens.disables)
	}
}

func TestVendorSupervisorRetainsActiveCollectorWhenSettingsUnavailable(t *testing.T) {
	tokens := &fakeTokenLifecycle{}
	replacements := 0
	s := newVendorTestSupervisor(tokens)
	s.replaceChildFn = func(context.Context, string, map[string]string) error {
		replacements++
		return nil
	}
	s.fetchSettingsFn = func(context.Context) (vendorSettings, error) {
		return vendorSettings{}, errors.New("settings unavailable")
	}
	s.reconcile(context.Background(), vendorSettings{enabled: true, endpoint: "https://relay.example.com"})

	s.refreshSettings(context.Background())
	if !s.enabled || s.activeEndpoint != "https://relay.example.com" || replacements != 1 || tokens.disables != 0 {
		t.Fatalf("settings failure changed active configuration: enabled=%t active=%q replacements=%d disables=%d", s.enabled, s.activeEndpoint, replacements, tokens.disables)
	}
}

func TestVendorSupervisorCrashRetainsTokenAndSchedulesRestart(t *testing.T) {
	tokens := &fakeTokenLifecycle{}
	s := newVendorTestSupervisor(tokens)
	s.enabled = true
	s.activeEndpoint = "https://relay.example.com"
	s.desiredEndpoint = s.activeEndpoint
	child := &childProcess{done: make(chan struct{}), startedAt: time.Now()}
	close(child.done)
	s.child = child
	s.stopChildFn = func() {
		s.mu.Lock()
		s.child = nil
		s.mu.Unlock()
	}

	s.restartIfNeeded(context.Background())
	if s.enabled || s.activeEndpoint != "" || s.desiredEndpoint != "https://relay.example.com" || s.nextStart.IsZero() || tokens.disables != 0 {
		t.Fatalf("collector crash discarded restart state or token: enabled=%t active=%q desired=%q next=%s disables=%d", s.enabled, s.activeEndpoint, s.desiredEndpoint, s.nextStart, tokens.disables)
	}
}

func TestVendorSupervisorRetriesTokenFailureWithoutStartingCollector(t *testing.T) {
	tokens := &fakeTokenLifecycle{enableErr: errors.New("issuer unavailable")}
	replacements := 0
	s := newVendorTestSupervisor(tokens)
	s.desiredEndpoint = "https://relay.example.com"
	s.replaceChildFn = func(context.Context, string, map[string]string) error {
		replacements++
		return nil
	}

	s.startCollector(context.Background())
	if replacements != 0 || s.enabled || s.nextStart.IsZero() {
		t.Fatalf("token failure activated vendor collector: replacements=%d next=%s", replacements, s.nextStart)
	}
}

func TestVendorSupervisorBackoffSurvivesShortLivedRestarts(t *testing.T) {
	s := newVendorTestSupervisor(&fakeTokenLifecycle{})
	s.replaceChildFn = func(context.Context, string, map[string]string) error {
		s.child = &childProcess{done: make(chan struct{}), startedAt: time.Now()}
		return nil
	}
	s.stopChildFn = func() { s.child = nil }
	s.reconcile(context.Background(), vendorSettings{enabled: true, endpoint: "https://relay.example.com"})

	for _, expected := range []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second} {
		close(s.child.done)
		s.restartIfNeeded(context.Background())
		if s.backoff != expected || s.nextStart.IsZero() {
			t.Fatalf("short-lived crash did not increase backoff: got %s, want %s", s.backoff, expected)
		}
		s.nextStart = time.Now().Add(-time.Second)
		s.restartIfNeeded(context.Background())
		if !s.enabled || s.child == nil || !s.nextStart.IsZero() {
			t.Fatal("scheduled collector restart did not succeed")
		}
	}

	s.child.startedAt = time.Now().Add(-time.Minute)
	close(s.child.done)
	s.restartIfNeeded(context.Background())
	if s.backoff != 2*time.Second {
		t.Fatalf("stable collector did not reset restart backoff: %s", s.backoff)
	}
}

func TestVendorSupervisorStopsTokenWhenInitialCollectorStartFails(t *testing.T) {
	tokens := &fakeTokenLifecycle{}
	s := newVendorTestSupervisor(tokens)
	s.desiredEndpoint = "https://relay.example.com"
	s.replaceChildFn = func(context.Context, string, map[string]string) error {
		return errors.New("collector failed")
	}

	s.startCollector(context.Background())
	if s.enabled || tokens.enables != 1 || tokens.disables != 1 || s.nextStart.IsZero() {
		t.Fatalf("failed collector retained token lifecycle: enabled=%t enables=%d disables=%d", s.enabled, tokens.enables, tokens.disables)
	}
}

func TestVendorSupervisorRunRequiresInstallAndNonLocalRunner(t *testing.T) {
	tests := map[string]*VendorSupervisor{
		"missing install": newVendorTestSupervisor(&fakeTokenLifecycle{}),
		"local runner":    newVendorTestSupervisor(&fakeTokenLifecycle{}),
	}
	tests["missing install"].installID = ""
	tests["local runner"].installID = "inst-test"
	tests["local runner"].local = true
	for name, supervisor := range tests {
		t.Run(name, func(t *testing.T) {
			supervisor.run(context.Background())
			select {
			case <-supervisor.done:
			default:
				t.Fatal("ineligible vendor supervisor did not exit")
			}
			if supervisor.tokens.(*fakeTokenLifecycle).disables != 1 {
				t.Fatalf("ineligible vendor supervisor did not remove stale token state: disables=%d", supervisor.tokens.(*fakeTokenLifecycle).disables)
			}
		})
	}
}

func TestVendorSupervisorShutdownCancelsReplacementWithoutRollback(t *testing.T) {
	tokens := &fakeTokenLifecycle{}
	s := newVendorTestSupervisor(tokens)
	s.enabled = true
	s.activeEndpoint = "https://previous.example.com"
	s.desiredEndpoint = s.activeEndpoint
	s.initialSettings = vendorSettings{enabled: true, endpoint: "https://replacement.example.com"}
	started := make(chan struct{})
	replacements, stops := 0, 0
	s.replaceChildFn = func(ctx context.Context, _ string, _ map[string]string) error {
		replacements++
		if replacements == 1 {
			close(started)
		}
		<-ctx.Done()
		return ctx.Err()
	}
	s.stopChildFn = func() { stops++ }
	if err := s.start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer s.cancel()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("replacement did not start")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.stop(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case <-s.done:
	default:
		t.Fatal("shutdown returned while replacement was still running")
	}
	if ctx.Err() != nil || replacements != 1 || stops != 1 || tokens.disables == 0 {
		t.Fatalf("shutdown retried replacement or exhausted its deadline: err=%v replacements=%d stops=%d disables=%d", ctx.Err(), replacements, stops, tokens.disables)
	}
}

func TestVendorSupervisorDoesNotStartAfterCancellation(t *testing.T) {
	tokens := &fakeTokenLifecycle{}
	s := newVendorTestSupervisor(tokens)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.startCollector(ctx)
	if tokens.enables != 0 {
		t.Fatal("canceled supervisor requested a token")
	}
	if err := s.replaceChild(ctx, "https://relay.example.com", nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled replacement accessed collector resources: %v", err)
	}
}

func TestVendorSupervisorRefreshesAttributeSnapshots(t *testing.T) {
	s := newVendorTestSupervisor(&fakeTokenLifecycle{})
	var applied []map[string]string
	s.replaceChildFn = func(_ context.Context, _ string, attributes map[string]string) error {
		applied = append(applied, maps.Clone(attributes))
		return nil
	}
	current := vendorSettings{enabled: true, endpoint: "https://relay.example.com", attributes: map[string]string{
		"nuon.install.name": "before", "nuon.install.labels.tier": "standard",
	}}
	s.fetchSettingsFn = func(context.Context) (vendorSettings, error) { return current, nil }
	s.refreshSettings(context.Background())
	current.attributes = map[string]string{"nuon.install.labels.tier": "standard", "nuon.install.name": "before"}
	s.refreshSettings(context.Background())
	if len(applied) != 1 {
		t.Fatal("identical attribute contents restarted the collector")
	}
	current.attributes["nuon.install.name"] = "after"
	delete(current.attributes, "nuon.install.labels.tier")
	if s.activeAttributes["nuon.install.name"] != "before" || s.desiredAttributes["nuon.install.name"] != "before" {
		t.Fatal("settings mutation modified an owned snapshot")
	}
	s.refreshSettings(context.Background())
	if len(applied) != 2 || !maps.Equal(applied[1], map[string]string{"nuon.install.name": "after"}) {
		t.Fatalf("rename and label deletion were not applied: %#v", applied)
	}
	s.fetchSettingsFn = func(context.Context) (vendorSettings, error) { return vendorSettings{}, errors.New("unavailable") }
	s.refreshSettings(context.Background())
	if !s.enabled || len(applied) != 2 || !maps.Equal(s.activeAttributes, applied[1]) {
		t.Fatal("settings failure discarded active attributes")
	}
	s.disable()
	if s.activeAttributes != nil || s.desiredAttributes != nil {
		t.Fatal("disabled collector retained metadata")
	}
}

func TestVendorSupervisorRollsBackAndRetriesAttributeChange(t *testing.T) {
	s := newVendorTestSupervisor(&fakeTokenLifecycle{})
	before := map[string]string{"nuon.install.name": "before", "nuon.install.labels.tier": "standard"}
	after := map[string]string{"nuon.install.name": "after"}
	var applied []map[string]string
	fail := true
	s.replaceChildFn = func(_ context.Context, endpoint string, attributes map[string]string) error {
		if endpoint != "https://relay.example.com" {
			t.Fatalf("metadata change altered endpoint: %s", endpoint)
		}
		applied = append(applied, maps.Clone(attributes))
		if fail && attributes["nuon.install.name"] == "after" {
			return errors.New("replacement failed")
		}
		return nil
	}
	s.reconcile(context.Background(), vendorSettings{enabled: true, endpoint: "https://relay.example.com", attributes: before})
	update := vendorSettings{enabled: true, endpoint: "https://relay.example.com", attributes: after}
	s.reconcile(context.Background(), update)
	if len(applied) != 3 || !maps.Equal(applied[2], before) || !maps.Equal(s.activeAttributes, before) || !maps.Equal(s.desiredAttributes, after) || !s.enabled || s.nextStart.IsZero() {
		t.Fatalf("metadata-only failure did not roll back the complete snapshot: %#v", applied)
	}
	s.reconcile(context.Background(), update)
	if len(applied) != 3 {
		t.Fatal("unchanged failed snapshot bypassed retry backoff")
	}
	fail = false
	s.nextStart = time.Now().Add(-time.Second)
	s.restartIfNeeded(context.Background())
	if len(applied) != 4 || !maps.Equal(applied[3], after) || !maps.Equal(s.activeAttributes, after) || !s.nextStart.IsZero() {
		t.Fatalf("retry failed to apply the desired attributes: %#v", applied)
	}
}

func newVendorTestSupervisor(tokens tokenLifecycle) *VendorSupervisor {
	s := &VendorSupervisor{
		logger:    zap.NewNop(),
		tokens:    tokens,
		done:      make(chan struct{}),
		backoff:   time.Second,
		installID: "inst-test",
	}
	s.fetchSettingsFn = func(context.Context) (vendorSettings, error) {
		return vendorSettings{}, nil
	}
	s.replaceChildFn = func(context.Context, string, map[string]string) error { return nil }
	s.stopChildFn = func() {}
	return s
}
