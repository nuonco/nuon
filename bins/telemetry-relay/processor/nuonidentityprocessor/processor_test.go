package nuonidentityprocessor

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"

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
	attributes.PutStr("nuon.install.id", "forged")
	attributes.PutStr("nuon.untrusted", "forged")
	attributes.PutStr("nuon_install_id", "forged")
	attributes.PutStr("Nuon_Org_ID", "forged")
}

func requireOnlyUnreservedAttributes(t *testing.T, attributes pcommon.Map) {
	t.Helper()
	attributes.Range(func(key string, _ pcommon.Value) bool {
		key = strings.ToLower(key)
		require.False(t, strings.HasPrefix(key, reservedAttributePrefix), key)
		require.False(t, strings.HasPrefix(key, normalizedReservedAttributePrefix), key)
		return true
	})
	requireAttribute(t, attributes, "keep", "value")
}

func requireStampedResource(t *testing.T, attributes pcommon.Map) {
	t.Helper()
	principal := identityTestPrincipal()
	requireAttribute(t, attributes, "keep", "value")
	requireAttribute(t, attributes, "nuon.org.id", principal.OrgID)
	requireAttribute(t, attributes, "nuon.app.id", principal.AppID)
	requireAttribute(t, attributes, "nuon.install.id", principal.InstallID)
	requireAttribute(t, attributes, "nuon.runner.id", principal.RunnerID)
	require.Equal(t, 5, attributes.Len())
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

	require.NoError(t, processLogs(identityTestContext(), logs))

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

	require.NoError(t, processTraces(identityTestContext(), traces))

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

	require.NoError(t, processMetrics(identityTestContext(), metrics))

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
	require.ErrorIs(t, processLogs(context.Background(), plog.NewLogs()), errMissingPrincipal)
	require.ErrorIs(t, processMetrics(context.Background(), pmetric.NewMetrics()), errMissingPrincipal)
	require.ErrorIs(t, processTraces(context.Background(), ptrace.NewTraces()), errMissingPrincipal)
}
