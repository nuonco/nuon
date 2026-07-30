package consumer

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	pkgkafka "github.com/nuonco/nuon/pkg/kafka"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

// TestOtelTraceEnvelopeRoundTrip asserts that a span survives the envelope
// unchanged, field for field, so the row this consumer inserts is the row the
// inline write path would have inserted.
//
// Reflective rather than a list of assertions on purpose: a new column added to
// OtelTraceIngestion without a json tag is exactly the regression this guards,
// and an explicit field list would keep passing while silently ignoring it.
//
// The nested Events*/Links* fields are the reason this exists. They carried
// `json:"-"` when they were only ever written by GORM, which would have dropped
// every span's events and links on the Kafka path while leaving the inline path
// correct — a divergence visible only in ClickHouse, long after the fact.
//
// Envelope-level rather than a call into the consumer: Decode's other behaviours
// (type filtering, dead-lettering) belong to the runtime package, and what's
// domain-specific here is only whether this payload survives the trip.
func TestOtelTraceEnvelopeRoundTrip(t *testing.T) {
	ts := time.Date(2026, 7, 30, 1, 2, 3, 123456789, time.UTC)

	in := app.OtelTraceIngestion{
		ID:                     "otltest",
		CreatedByID:            "acct_1",
		CreatedAt:              ts,
		UpdatedAt:              ts,
		RunnerID:               "rnr_1",
		RunnerJobID:            "job_1",
		RunnerGroupID:          "grp_1",
		RunnerJobExecutionID:   "exec_1",
		RunnerJobExecutionStep: "apply",
		Timestamp:              ts,
		TimestampDate:          ts,
		TimestampTime:          ts,
		ResourceAttributes:     map[string]string{"runner_group.id": "grp_1"},
		ResourceSchemaURL:      "https://schema/res",
		ScopeName:              "runner.helm",
		ScopeVersion:           "v1",
		ScopeSchemaURL:         "https://schema/scope",
		ScopeAttributes:        map[string]string{"a": "b"},
		ScopeDroppedAttrCount:  2,
		TraceID:                "abc",
		SpanID:                 "def",
		ParentSpanID:           "ghi",
		TraceState:             "x=1",
		SpanName:               "terraform-apply",
		SpanKind:               "SPAN_KIND_INTERNAL",
		ServiceName:            "runner",
		SpanAttributes:         map[string]string{"runner_job.id": "job_1"},
		Duration:               1500,
		StatusCode:             "STATUS_CODE_OK",
		StatusMessage:          "ok",
		EventsTimestamp:        []time.Time{ts, ts.Add(time.Second)},
		EventsName:             []string{"start", "end"},
		EventsAttributes:       []map[string]string{{"k": "v"}, {}},
		LinksTraceID:           []string{"lt1"},
		LinksSpanID:            []string{"ls1"},
		LinksState:             []string{"ls=1"},
		LinksAttributes:        []map[string]string{{"lk": "lv"}},
	}

	raw, err := pkgkafka.Wrap("test", kafka.TypeOtelTrace, in)
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}

	env, err := pkgkafka.Unwrap(raw)
	if err != nil {
		t.Fatalf("unwrap: %v", err)
	}

	var out app.OtelTraceIngestion
	if err := json.Unmarshal(env.Payload, &out); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	typ := reflect.TypeOf(in)
	inVal, outVal := reflect.ValueOf(in), reflect.ValueOf(out)
	for i := range typ.NumField() {
		field := typ.Field(i)

		// DeletedAt is the one field the write path never sets and the table
		// defaults, so it is deliberately excluded from the payload.
		if field.Name == "DeletedAt" {
			continue
		}

		got, want := outVal.Field(i).Interface(), inVal.Field(i).Interface()
		if !reflect.DeepEqual(want, got) {
			t.Errorf("%s not preserved through the envelope:\n  produced: %#v\n  consumed: %#v", field.Name, want, got)
		}
	}
}
