package nuonidentityprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/nuonco/nuon/bins/telemetry-relay/extension/nuonjwtauthextension"
)

func identityTestPrincipal() nuonjwtauthextension.Principal {
	return nuonjwtauthextension.Principal{
		OrgID:     "org-test",
		AppID:     "app-test",
		InstallID: "install-test",
		RunnerID:  "runner-test",
	}
}

func identityTestContext() context.Context {
	info := client.FromContext(context.Background())
	info.Auth = nuonjwtauthextension.NewAuthData(identityTestPrincipal())
	return client.NewContext(context.Background(), info)
}

func addUntrustedAttributes(attributes pcommon.Map) {
	attributes.PutStr("keep", "value")
	for _, key := range []string{
		"nuon.org.id", "nuon.app.id", "nuon.install.id", "nuon.runner.id",
		"nuon_org_id", "nuon_app_id", "nuon_install_id", "nuon_runner_id",
		"Nuon_Org_ID", "NUON.Runner.ID",
	} {
		attributes.PutStr(key, "forged")
	}
	attributes.PutStr("nuon.untrusted", "forged")
	attributes.PutStr("nuon.org.name", "example-org")
	attributes.PutStr("nuon.app.name", "payments")
	attributes.PutStr("nuon.install.name", "example-install")
	attributes.PutStr("nuon.runner.name", "example-runner")
	attributes.PutStr("nuon_org_name", "example-org")
	attributes.PutStr("nuon.api", "public")
	attributes.PutStr("nuon.custom.id", "custom-id")
	attributes.PutStr("nuon.org.id.extra", "extra")
	attributes.PutStr("nuon_org_id_extra", "extra")
}

func requireOnlyUnreservedAttributes(t *testing.T, attributes pcommon.Map) {
	t.Helper()
	require.Equal(t, map[string]any{
		"keep":              "value",
		"nuon.untrusted":    "forged",
		"nuon.org.name":     "example-org",
		"nuon.app.name":     "payments",
		"nuon.install.name": "example-install",
		"nuon.runner.name":  "example-runner",
		"nuon_org_name":     "example-org",
		"nuon.api":          "public",
		"nuon.custom.id":    "custom-id",
		"nuon.org.id.extra": "extra",
		"nuon_org_id_extra": "extra",
	}, attributes.AsRaw())
}

func requireStampedResource(t *testing.T, attributes pcommon.Map) {
	t.Helper()
	principal := identityTestPrincipal()
	requireAttribute(t, attributes, "keep", "value")
	requireAttribute(t, attributes, "nuon.org.id", principal.OrgID)
	requireAttribute(t, attributes, "nuon.app.id", principal.AppID)
	requireAttribute(t, attributes, "nuon.install.id", principal.InstallID)
	requireAttribute(t, attributes, "nuon.runner.id", principal.RunnerID)
	preserved := pcommon.NewMap()
	attributes.CopyTo(preserved)
	for _, key := range []string{"nuon.org.id", "nuon.app.id", "nuon.install.id", "nuon.runner.id"} {
		preserved.Remove(key)
	}
	requireOnlyUnreservedAttributes(t, preserved)
}

func requireAttribute(t *testing.T, attributes pcommon.Map, key, expected string) {
	t.Helper()
	value, ok := attributes.Get(key)
	require.True(t, ok)
	require.Equal(t, expected, value.Str())
}

func TestProcessLogsStripsReservedAttributesAndStampsResources(t *testing.T) {
	logs := plog.NewLogs()
	resourceLogs := logs.ResourceLogs().AppendEmpty()
	addUntrustedAttributes(resourceLogs.Resource().Attributes())
	scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
	addUntrustedAttributes(scopeLogs.Scope().Attributes())
	record := scopeLogs.LogRecords().AppendEmpty()
	addUntrustedAttributes(record.Attributes())

	require.NoError(t, processLogs(identityTestContext(), logs, nil))

	requireStampedResource(t, resourceLogs.Resource().Attributes())
	requireOnlyUnreservedAttributes(t, scopeLogs.Scope().Attributes())
	requireOnlyUnreservedAttributes(t, record.Attributes())
}

func TestProcessTracesStripsReservedAttributesAndStampsResources(t *testing.T) {
	traces := ptrace.NewTraces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()
	addUntrustedAttributes(resourceSpans.Resource().Attributes())
	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty()
	addUntrustedAttributes(scopeSpans.Scope().Attributes())
	span := scopeSpans.Spans().AppendEmpty()
	addUntrustedAttributes(span.Attributes())
	event := span.Events().AppendEmpty()
	addUntrustedAttributes(event.Attributes())
	link := span.Links().AppendEmpty()
	addUntrustedAttributes(link.Attributes())

	require.NoError(t, processTraces(identityTestContext(), traces, nil))

	requireStampedResource(t, resourceSpans.Resource().Attributes())
	requireOnlyUnreservedAttributes(t, scopeSpans.Scope().Attributes())
	requireOnlyUnreservedAttributes(t, span.Attributes())
	requireOnlyUnreservedAttributes(t, event.Attributes())
	requireOnlyUnreservedAttributes(t, link.Attributes())
}

func TestProcessMetricsStripsEveryMetricAttributeLocation(t *testing.T) {
	metrics := pmetric.NewMetrics()
	resourceMetrics := metrics.ResourceMetrics().AppendEmpty()
	addUntrustedAttributes(resourceMetrics.Resource().Attributes())
	scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
	addUntrustedAttributes(scopeMetrics.Scope().Attributes())

	gauge := scopeMetrics.Metrics().AppendEmpty()
	gauge.SetName("gauge")
	gaugePoint := gauge.SetEmptyGauge().DataPoints().AppendEmpty()
	addUntrustedAttributes(gauge.Metadata())
	addUntrustedAttributes(gaugePoint.Attributes())
	addUntrustedAttributes(gaugePoint.Exemplars().AppendEmpty().FilteredAttributes())

	sum := scopeMetrics.Metrics().AppendEmpty()
	sum.SetName("sum")
	sumPoint := sum.SetEmptySum().DataPoints().AppendEmpty()
	addUntrustedAttributes(sum.Metadata())
	addUntrustedAttributes(sumPoint.Attributes())
	addUntrustedAttributes(sumPoint.Exemplars().AppendEmpty().FilteredAttributes())

	histogram := scopeMetrics.Metrics().AppendEmpty()
	histogram.SetName("histogram")
	histogramPoint := histogram.SetEmptyHistogram().DataPoints().AppendEmpty()
	addUntrustedAttributes(histogram.Metadata())
	addUntrustedAttributes(histogramPoint.Attributes())
	addUntrustedAttributes(histogramPoint.Exemplars().AppendEmpty().FilteredAttributes())

	exponentialHistogram := scopeMetrics.Metrics().AppendEmpty()
	exponentialHistogram.SetName("exponential-histogram")
	exponentialPoint := exponentialHistogram.SetEmptyExponentialHistogram().DataPoints().AppendEmpty()
	addUntrustedAttributes(exponentialHistogram.Metadata())
	addUntrustedAttributes(exponentialPoint.Attributes())
	addUntrustedAttributes(exponentialPoint.Exemplars().AppendEmpty().FilteredAttributes())

	summary := scopeMetrics.Metrics().AppendEmpty()
	summary.SetName("summary")
	summaryPoint := summary.SetEmptySummary().DataPoints().AppendEmpty()
	addUntrustedAttributes(summary.Metadata())
	addUntrustedAttributes(summaryPoint.Attributes())

	empty := scopeMetrics.Metrics().AppendEmpty()
	empty.SetName("empty")
	addUntrustedAttributes(empty.Metadata())

	require.NoError(t, processMetrics(identityTestContext(), metrics, nil))

	requireStampedResource(t, resourceMetrics.Resource().Attributes())
	requireOnlyUnreservedAttributes(t, scopeMetrics.Scope().Attributes())
	for _, attributes := range []pcommon.Map{
		gauge.Metadata(), gaugePoint.Attributes(), gaugePoint.Exemplars().At(0).FilteredAttributes(),
		sum.Metadata(), sumPoint.Attributes(), sumPoint.Exemplars().At(0).FilteredAttributes(),
		histogram.Metadata(), histogramPoint.Attributes(), histogramPoint.Exemplars().At(0).FilteredAttributes(),
		exponentialHistogram.Metadata(), exponentialPoint.Attributes(), exponentialPoint.Exemplars().At(0).FilteredAttributes(),
		summary.Metadata(), summaryPoint.Attributes(), empty.Metadata(),
	} {
		requireOnlyUnreservedAttributes(t, attributes)
	}
}

func TestProcessorsRequireVerifiedPrincipal(t *testing.T) {
	require.ErrorIs(t, processLogs(context.Background(), plog.NewLogs(), nil), errMissingPrincipal)
	require.ErrorIs(t, processMetrics(context.Background(), pmetric.NewMetrics(), nil), errMissingPrincipal)
	require.ErrorIs(t, processTraces(context.Background(), ptrace.NewTraces(), nil), errMissingPrincipal)
}

func TestFactoryEnforcesAllowedOrgs(t *testing.T) {
	for _, tc := range []struct {
		name    string
		orgIDs  []string
		allowed bool
	}{
		{name: "unset allows all", allowed: true},
		{name: "empty allows all", orgIDs: []string{}, allowed: true},
		{name: "single matching org", orgIDs: []string{"org-test"}, allowed: true},
		{name: "match later in list", orgIDs: []string{"org-other", "org-test"}, allowed: true},
		{name: "payload org cannot grant access", orgIDs: []string{"forged"}},
		{name: "exact match required", orgIDs: []string{"org", "ORG-TEST", "org-test-extra"}},
	} {
		for _, signal := range []string{"logs", "metrics", "traces"} {
			t.Run(tc.name+"/"+signal, func(t *testing.T) {
				cfg := &Config{AllowedOrgIDs: tc.orgIDs}
				settings := processortest.NewNopSettings(componentType)
				factory := NewFactory()
				ctx := identityTestContext()
				forwarded := 0
				var attributes pcommon.Map
				var consumeErr error

				switch signal {
				case "logs":
					logs := plog.NewLogs()
					rsrc := logs.ResourceLogs().AppendEmpty()
					attributes = rsrc.Resource().Attributes()
					addUntrustedAttributes(attributes)
					rsrc.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStr("example log")
					sink, err := consumer.NewLogs(func(_ context.Context, data plog.Logs) error {
						forwarded += data.LogRecordCount()
						requireStampedResource(t, data.ResourceLogs().At(0).Resource().Attributes())
						return nil
					})
					require.NoError(t, err)
					proc, err := factory.CreateLogs(context.Background(), settings, cfg, sink)
					require.NoError(t, err)
					t.Cleanup(func() { require.NoError(t, proc.Shutdown(context.Background())) })
					consumeErr = proc.ConsumeLogs(ctx, logs)
				case "metrics":
					metrics := pmetric.NewMetrics()
					rsrc := metrics.ResourceMetrics().AppendEmpty()
					attributes = rsrc.Resource().Attributes()
					addUntrustedAttributes(attributes)
					metric := rsrc.ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
					metric.SetName("example.gauge")
					metric.SetEmptyGauge().DataPoints().AppendEmpty().SetIntValue(7)
					sink, err := consumer.NewMetrics(func(_ context.Context, data pmetric.Metrics) error {
						forwarded += data.DataPointCount()
						requireStampedResource(t, data.ResourceMetrics().At(0).Resource().Attributes())
						return nil
					})
					require.NoError(t, err)
					proc, err := factory.CreateMetrics(context.Background(), settings, cfg, sink)
					require.NoError(t, err)
					t.Cleanup(func() { require.NoError(t, proc.Shutdown(context.Background())) })
					consumeErr = proc.ConsumeMetrics(ctx, metrics)
				case "traces":
					traces := ptrace.NewTraces()
					rsrc := traces.ResourceSpans().AppendEmpty()
					attributes = rsrc.Resource().Attributes()
					addUntrustedAttributes(attributes)
					rsrc.ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("example span")
					sink, err := consumer.NewTraces(func(_ context.Context, data ptrace.Traces) error {
						forwarded += data.SpanCount()
						requireStampedResource(t, data.ResourceSpans().At(0).Resource().Attributes())
						return nil
					})
					require.NoError(t, err)
					proc, err := factory.CreateTraces(context.Background(), settings, cfg, sink)
					require.NoError(t, err)
					t.Cleanup(func() { require.NoError(t, proc.Shutdown(context.Background())) })
					consumeErr = proc.ConsumeTraces(ctx, traces)
				}

				if tc.allowed {
					require.NoError(t, consumeErr)
					require.Equal(t, 1, forwarded)
				} else {
					require.ErrorIs(t, consumeErr, errOrgNotAllowed)
					require.True(t, consumererror.IsPermanent(consumeErr))
					require.Zero(t, forwarded)
					requireAttribute(t, attributes, "nuon.org.id", "forged")
				}
			})
		}
	}
}

func TestConfigValidateAllowedOrgIDs(t *testing.T) {
	require.NoError(t, createDefaultConfig().(*Config).Validate())
	require.NoError(t, (&Config{AllowedOrgIDs: []string{"org-first", "org-second"}}).Validate())
	for _, orgID := range []string{"", " ", " org-test", "org-test\n"} {
		cfg := &Config{AllowedOrgIDs: []string{"org-test", orgID}}
		require.ErrorContains(t, cfg.Validate(), "allowed_org_ids[1]")
	}
}
