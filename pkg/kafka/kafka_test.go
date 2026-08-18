package kafka

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
)

type testPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestWrapUnwrapRoundTrip(t *testing.T) {
	orig := testPayload{ID: "abc123", Name: "widget"}

	b, err := Wrap("my-source", "widget.created", orig)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	env, err := Unwrap(b)
	require.NoError(t, err)

	require.Equal(t, EnvelopeVersion, env.Version)
	require.Equal(t, "widget.created", env.Type)
	require.Equal(t, "my-source", env.Source)
	require.False(t, env.ProducedAt.IsZero())

	var got testPayload
	require.NoError(t, json.Unmarshal(env.Payload, &got))
	require.Equal(t, orig, got)
}

func TestUnwrap_VersionGuard(t *testing.T) {
	t.Run("newer than supported is rejected", func(t *testing.T) {
		env := Envelope{Version: EnvelopeVersion + 1, Type: "widget.created", Payload: json.RawMessage(`{}`)}
		b, err := json.Marshal(env)
		require.NoError(t, err)

		_, err = Unwrap(b)
		require.Error(t, err)
	})

	t.Run("equal to supported is accepted", func(t *testing.T) {
		env := Envelope{Version: EnvelopeVersion, Type: "widget.created", Payload: json.RawMessage(`{}`)}
		b, err := json.Marshal(env)
		require.NoError(t, err)

		got, err := Unwrap(b)
		require.NoError(t, err)
		require.Equal(t, EnvelopeVersion, got.Version)
	})

	t.Run("older than supported is accepted", func(t *testing.T) {
		env := Envelope{Version: EnvelopeVersion - 1, Type: "widget.created", Payload: json.RawMessage(`{}`)}
		b, err := json.Marshal(env)
		require.NoError(t, err)

		got, err := Unwrap(b)
		require.NoError(t, err)
		require.Equal(t, EnvelopeVersion-1, got.Version)
	})
}

func TestWrapUnwrapRoundTrip_MapPayload(t *testing.T) {
	orig := map[string]any{"foo": "bar", "count": float64(3)}

	b, err := Wrap("map-source", "map.event", orig)
	require.NoError(t, err)

	env, err := Unwrap(b)
	require.NoError(t, err)

	require.Equal(t, EnvelopeVersion, env.Version)
	require.Equal(t, "map.event", env.Type)
	require.Equal(t, "map-source", env.Source)
	require.False(t, env.ProducedAt.IsZero())

	var got map[string]any
	require.NoError(t, json.Unmarshal(env.Payload, &got))
	require.Equal(t, orig, got)
}

func TestConfigBaseOpts(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "no brokers",
			cfg:     Config{},
			wantErr: true,
		},
		{
			name: "unset protocol defaults to plaintext",
			cfg: Config{
				Brokers: []string{"localhost:9092"},
			},
			wantErr: false,
		},
		{
			name: "explicit plaintext",
			cfg: Config{
				Brokers:          []string{"localhost:9092"},
				SecurityProtocol: securityPlaintext,
			},
			wantErr: false,
		},
		{
			name: "lowercase plaintext",
			cfg: Config{
				Brokers:          []string{"localhost:9092"},
				SecurityProtocol: "plaintext",
			},
			wantErr: false,
		},
		{
			name: "sasl_ssl is not supported",
			cfg: Config{
				Brokers:          []string{"broker:9093"},
				SecurityProtocol: "SASL_SSL",
			},
			wantErr: true,
		},
		{
			name: "unknown protocol",
			cfg: Config{
				Brokers:          []string{"broker:9092"},
				SecurityProtocol: "NOPE",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := tt.cfg.baseOpts(zap.NewNop())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, opts)
		})
	}
}

func TestDedupToken(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		partition int32
		first     int64
		last      int64
		want      string
	}{
		{
			name:      "basic",
			topic:     "otel_log_records",
			partition: 3,
			first:     1000,
			last:      1999,
			want:      "otel_log_records:3:1000-1999",
		},
		{
			name:      "zero partition",
			topic:     "events",
			partition: 0,
			first:     0,
			last:      10,
			want:      "events:0:0-10",
		},
		{
			name:      "single record range",
			topic:     "spans",
			partition: 7,
			first:     42,
			last:      42,
			want:      "spans:7:42-42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DedupToken(tt.topic, tt.partition, tt.first, tt.last)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestKafkaBatchTracePropagation(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { require.NoError(t, tp.Shutdown(context.Background())) })

	kafkaTracer := kotel.NewTracer(
		kotel.TracerProvider(tp),
		kotel.TracerPropagator(propagation.TraceContext{}),
	)

	var requestSpans []trace.Span
	var requestSpanContexts []trace.SpanContext
	var publishSpanContexts []trace.SpanContext
	var receiveSpanContexts []trace.SpanContext
	var records []*kgo.Record
	for i, key := range []string{"first", "second"} {
		requestCtx, requestSpan := tp.Tracer("test").Start(context.Background(), "request "+key)
		requestSpans = append(requestSpans, requestSpan)
		requestSpanContexts = append(requestSpanContexts, requestSpan.SpanContext())

		produced := &kgo.Record{Topic: "events", Key: []byte(key), Context: requestCtx}
		kafkaTracer.OnProduceRecordBuffered(produced)
		require.NotEmpty(t, produced.Headers)
		publishSpanContexts = append(publishSpanContexts, trace.SpanContextFromContext(produced.Context))
		kafkaTracer.OnProduceRecordUnbuffered(produced, nil)

		consumed := &kgo.Record{
			Topic:     produced.Topic,
			Key:       produced.Key,
			Headers:   produced.Headers,
			Partition: 2,
			Offset:    int64(42 + i),
			Context:   context.Background(),
		}
		kafkaTracer.OnFetchRecordBuffered(consumed)
		receiveSpanContexts = append(receiveSpanContexts, trace.SpanContextFromContext(consumed.Context))
		kafkaTracer.OnFetchRecordUnbuffered(consumed, false)
		records = append(records, consumed)
	}

	var handlerSpanContext trace.SpanContext
	consumer := Consumer{
		batchTracer: tp.Tracer("test-batch"),
		handler: func(ctx context.Context, _ int32, _ []*kgo.Record) error {
			handlerSpanContext = trace.SpanContextFromContext(ctx)
			_, insertSpan := tp.Tracer("test").Start(ctx, "clickhouse insert")
			insertSpan.End()
			return nil
		},
	}
	require.NoError(t, consumer.handle(context.Background(), 2, records))
	for _, requestSpan := range requestSpans {
		requestSpan.End()
	}

	var batchSpan, insertSpan sdktrace.ReadOnlySpan
	for _, span := range recorder.Ended() {
		switch span.Name() {
		case "events process":
			batchSpan = span
		case "clickhouse insert":
			insertSpan = span
		}
	}
	require.NotNil(t, batchSpan)
	require.NotNil(t, insertSpan)
	require.False(t, batchSpan.Parent().IsValid())
	require.Equal(t, batchSpan.SpanContext(), handlerSpanContext)
	require.Equal(t, batchSpan.SpanContext().SpanID(), insertSpan.Parent().SpanID())
	require.Len(t, batchSpan.Links(), 2)
	require.Equal(t, requestSpanContexts[0].TraceID(), receiveSpanContexts[0].TraceID())
	require.Equal(t, requestSpanContexts[1].TraceID(), receiveSpanContexts[1].TraceID())
	require.NotEqual(t, receiveSpanContexts[0].TraceID(), batchSpan.SpanContext().TraceID())
	require.NotEqual(t, receiveSpanContexts[1].TraceID(), batchSpan.SpanContext().TraceID())
	require.ElementsMatch(t,
		[]trace.SpanID{publishSpanContexts[0].SpanID(), publishSpanContexts[1].SpanID()},
		[]trace.SpanID{batchSpan.Links()[0].SpanContext.SpanID(), batchSpan.Links()[1].SpanContext.SpanID()},
	)
}

func TestDisabledProducer(t *testing.T) {
	l := zap.NewNop()
	mw, err := metrics.New(validator.New(), metrics.WithDisable(true))
	require.NoError(t, err)

	p := DisabledProducer(l, mw)
	require.NotNil(t, p)
	require.False(t, p.Enabled())

	require.NoError(t, p.Ping(context.Background()))
	require.NoError(t, p.Flush(context.Background()))

	require.NotPanics(t, func() {
		p.Close()
	})
}
