package nuonidentityprocessor

import (
	"context"
	"errors"
	"strings"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/nuonco/nuon/bins/telemetry-relay/extension/nuonjwtauthextension"
)

const reservedAttributePrefix = "nuon."

var errMissingPrincipal = errors.New("verified Nuon telemetry principal is required")

func processLogs(ctx context.Context, logs plog.Logs) error {
	principal, err := principalFromContext(ctx)
	if err != nil {
		return err
	}

	resourceLogs := logs.ResourceLogs()
	for i := 0; i < resourceLogs.Len(); i++ {
		resourceLog := resourceLogs.At(i)
		stampResource(resourceLog.Resource().Attributes(), principal)
		scopeLogs := resourceLog.ScopeLogs()
		for j := 0; j < scopeLogs.Len(); j++ {
			scopeLog := scopeLogs.At(j)
			stripReserved(scopeLog.Scope().Attributes())
			records := scopeLog.LogRecords()
			for k := 0; k < records.Len(); k++ {
				stripReserved(records.At(k).Attributes())
			}
		}
	}
	return nil
}

func processMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	principal, err := principalFromContext(ctx)
	if err != nil {
		return err
	}

	resourceMetrics := metrics.ResourceMetrics()
	for i := 0; i < resourceMetrics.Len(); i++ {
		resourceMetric := resourceMetrics.At(i)
		stampResource(resourceMetric.Resource().Attributes(), principal)
		scopeMetrics := resourceMetric.ScopeMetrics()
		for j := 0; j < scopeMetrics.Len(); j++ {
			scopeMetric := scopeMetrics.At(j)
			stripReserved(scopeMetric.Scope().Attributes())
			metricSlice := scopeMetric.Metrics()
			for k := 0; k < metricSlice.Len(); k++ {
				stripMetric(metricSlice.At(k))
			}
		}
	}
	return nil
}

func processTraces(ctx context.Context, traces ptrace.Traces) error {
	principal, err := principalFromContext(ctx)
	if err != nil {
		return err
	}

	resourceSpans := traces.ResourceSpans()
	for i := 0; i < resourceSpans.Len(); i++ {
		resourceSpan := resourceSpans.At(i)
		stampResource(resourceSpan.Resource().Attributes(), principal)
		scopeSpans := resourceSpan.ScopeSpans()
		for j := 0; j < scopeSpans.Len(); j++ {
			scopeSpan := scopeSpans.At(j)
			stripReserved(scopeSpan.Scope().Attributes())
			spans := scopeSpan.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				stripReserved(span.Attributes())
				events := span.Events()
				for l := 0; l < events.Len(); l++ {
					stripReserved(events.At(l).Attributes())
				}
				links := span.Links()
				for l := 0; l < links.Len(); l++ {
					stripReserved(links.At(l).Attributes())
				}
			}
		}
	}
	return nil
}

func principalFromContext(ctx context.Context) (nuonjwtauthextension.Principal, error) {
	authData, ok := client.FromContext(ctx).Auth.(*nuonjwtauthextension.AuthData)
	if !ok {
		return nuonjwtauthextension.Principal{}, errMissingPrincipal
	}
	principal := authData.Principal()
	if principal.OrgID == "" || principal.AppID == "" || principal.InstallID == "" || principal.RunnerID == "" {
		return nuonjwtauthextension.Principal{}, errMissingPrincipal
	}
	return principal, nil
}

func stampResource(attributes pcommon.Map, principal nuonjwtauthextension.Principal) {
	stripReserved(attributes)
	attributes.PutStr("nuon.org.id", principal.OrgID)
	attributes.PutStr("nuon.app.id", principal.AppID)
	attributes.PutStr("nuon.install.id", principal.InstallID)
	attributes.PutStr("nuon.runner.id", principal.RunnerID)
}

func stripReserved(attributes pcommon.Map) {
	attributes.RemoveIf(func(key string, _ pcommon.Value) bool {
		return strings.HasPrefix(key, reservedAttributePrefix)
	})
}

func stripMetric(metric pmetric.Metric) {
	stripReserved(metric.Metadata())
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		stripNumberDataPoints(metric.Gauge().DataPoints())
	case pmetric.MetricTypeSum:
		stripNumberDataPoints(metric.Sum().DataPoints())
	case pmetric.MetricTypeHistogram:
		points := metric.Histogram().DataPoints()
		for i := 0; i < points.Len(); i++ {
			point := points.At(i)
			stripReserved(point.Attributes())
			stripExemplars(point.Exemplars())
		}
	case pmetric.MetricTypeExponentialHistogram:
		points := metric.ExponentialHistogram().DataPoints()
		for i := 0; i < points.Len(); i++ {
			point := points.At(i)
			stripReserved(point.Attributes())
			stripExemplars(point.Exemplars())
		}
	case pmetric.MetricTypeSummary:
		points := metric.Summary().DataPoints()
		for i := 0; i < points.Len(); i++ {
			stripReserved(points.At(i).Attributes())
		}
	case pmetric.MetricTypeEmpty:
	}
}

func stripNumberDataPoints(points pmetric.NumberDataPointSlice) {
	for i := 0; i < points.Len(); i++ {
		point := points.At(i)
		stripReserved(point.Attributes())
		stripExemplars(point.Exemplars())
	}
}

func stripExemplars(exemplars pmetric.ExemplarSlice) {
	for i := 0; i < exemplars.Len(); i++ {
		stripReserved(exemplars.At(i).FilteredAttributes())
	}
}
