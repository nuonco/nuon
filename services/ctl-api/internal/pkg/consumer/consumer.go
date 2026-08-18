// Package consumer is the Kafka consumer runtime: consumer naming and
// selection, the shared ClickHouse-writing sink, and the dead-letter path. It is
// to `consumer` what internal/pkg/worker is to `worker` — none of it knows
// anything about a particular domain.
//
// Domain consumers live with their domain (app/runners/consumer holds
// heartbeats, otel-logs and otel-traces) and consist of little more than a Sink
// plus a decode-and-insert handler. The one consumer that lives here is the DLQ
// (see dlq.go), because the dead-letter topic is shared infrastructure that
// every consumer produces to rather than any domain's data.
package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	chgo "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	pkgkafka "github.com/nuonco/nuon/pkg/kafka"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

// Name identifies one consumer: it is an element of the `consumer --name` value,
// the suffix of the consumer group, and the prefix of the consumer's metrics.
// One Name owns exactly one topic.
//
// A Name is NOT a deployment. A deployment (a pod, its resources, its
// SERVICE_DEPLOYMENT tag, and therefore its Kafka client id) may host more than
// one consumer — `--name=otel-logs,otel-traces` runs both in one pod under the
// deployment named `otel`, each still with its own group and its own client.
// Consumers sharing a pod share a client id, so they also share a Kafka client
// quota, and share a liveness probe: if one handler wedges, the restart takes
// the other down with it. Group them only when scaling and failing together is
// acceptable.
//
// The names live here, in the runtime, rather than each with its domain:
// NewSelection runs before the fx graph is built (see cmd/consumer.go), so the
// set of valid names has to be statically known and can't be collected from the
// domains at wiring time.
type Name string

const (
	NameHeartbeats Name = "heartbeats"
	NameOtelLogs   Name = "otel-logs"
	NameOtelTraces Name = "otel-traces"
	NameDLQ        Name = "dlq"
)

// NameAll selects every consumer, mirroring `worker --namespace=all`. Locally
// one process runs them all.
const NameAll = "all"

func Names() []Name {
	return []Name{NameHeartbeats, NameOtelLogs, NameOtelTraces, NameDLQ}
}

// Selection is the parsed `--name` flag.
type Selection struct {
	all   bool
	names map[Name]bool
}

// NewSelection parses a comma-separated list of consumer names, or "all".
//
// Unknown names are an error rather than a no-op: a typo in a deployment's flag
// would otherwise produce a pod that starts, passes health checks, and silently
// consumes nothing, which is a much harder failure to spot than a crash loop.
func NewSelection(spec string) (Selection, error) {
	s := Selection{names: map[Name]bool{}}

	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if part == NameAll {
			s.all = true
			continue
		}

		var match bool
		for _, known := range Names() {
			if Name(part) == known {
				s.names[known] = true
				match = true
				break
			}
		}
		if !match {
			return Selection{}, fmt.Errorf("unknown consumer %q: valid values are %s, or %q", part, Names(), NameAll)
		}
	}

	if !s.all && len(s.names) == 0 {
		return Selection{}, fmt.Errorf("no consumers selected: pass --name with one of %s, or %q", Names(), NameAll)
	}

	return s, nil
}

func (s Selection) Includes(n Name) bool {
	return s.all || s.names[n]
}

type Params struct {
	fx.In

	Cfg       *internal.Config
	L         *zap.Logger
	MW        metrics.Writer
	CHDB      *gorm.DB `name:"ch"`
	LC        fx.Lifecycle
	Selection Selection
	Producer  *kafka.Producer
}

// Sink is the shared half of every consumer that writes to ClickHouse: naming,
// metrics, the ClickHouse handle, and the poll-loop lifecycle. Each consumer
// embeds one and adds only its own decode-and-insert handler.
type Sink struct {
	name  Name
	topic string
	group string

	l        *zap.Logger
	mw       metrics.Writer
	chDB     *gorm.DB
	producer *kafka.Producer

	// writeTimeout bounds Insert(), so a wedged or degraded ClickHouse fails a
	// batch (for redelivery) rather than blocking this handler indefinitely.
	writeTimeout time.Duration
	// cons is set once Start() builds the poll loop, so a healthcheck can ask
	// whether this consumer's handler is currently stuck.
	cons *pkgkafka.Consumer

	// deadLetterFallbackOnly skips producing to the dead-letter topic and goes
	// straight to the direct-write fallback. Set only by the DLQ consumer, so a
	// dead-letter-about-a-dead-letter can't recurse through the topic.
	deadLetterFallbackOnly bool
}

// NewSink returns nil when this consumer should not run in this process — either
// it wasn't selected, or Kafka is disabled entirely. Callers return a nil
// consumer in that case and whatever inline write path they replaced stays in
// effect.
func NewSink(params Params, name Name, topic string) *Sink {
	l := params.L.Named("kafka-consumer-" + string(name))

	if !params.Selection.Includes(name) {
		l.Info("consumer not selected for this process; not started")
		return nil
	}
	if !params.Cfg.KafkaEnabled {
		l.Info("kafka disabled; consumer not started")
		return nil
	}

	return &Sink{
		name:         name,
		topic:        topic,
		group:        kafka.ConsumerGroup(params.Cfg, string(name)),
		l:            l,
		mw:           params.MW,
		chDB:         params.CHDB,
		producer:     params.Producer,
		writeTimeout: params.Cfg.ClickhouseDBWriteTimeout,
	}
}

// instrument wraps a domain handler with kafka.consumer.latency, timing the
// same span Stuck() already watches for liveness: decode, insert, and the
// dead-letter path if one gets triggered. Decode itself never fails the
// handler — only Insert can — so reason:insert_error is the only error reason
// this can currently produce.
func (s *Sink) instrument(handler pkgkafka.Handler) pkgkafka.Handler {
	return func(ctx context.Context, partition int32, recs []*kgo.Record) error {
		start := time.Now()

		// Seed a MetricContext so the GORM metrics plugin tags this consumer's
		// ClickHouse writes, the same way the Temporal activity interceptor does
		// for worker activities. Without it the plugin has no request scope to
		// attribute the write to and the per-table latency series is untagged.
		ctx = context.WithValue(ctx, keys.MetricsKey, &cctx.MetricContext{
			Endpoint:  string(s.name),
			Method:    "kafka",
			Context:   "consumer",
			Namespace: s.topic,
		})

		err := handler(ctx, partition, recs)

		tags := s.baseTags("status:ok")
		if err != nil {
			tags = s.baseTags("status:err", "reason:insert_error")
		}
		s.mw.Timing("kafka.consumer.latency", time.Since(start), tags)
		return err
	}
}

// Start builds the consumer client and binds its poll loop to the fx lifecycle.
func (s *Sink) Start(params Params, handler pkgkafka.Handler) error {
	cons, err := pkgkafka.NewConsumer(
		kafka.ClientConfig(params.Cfg),
		kafka.ConsumerConfig(params.Cfg, string(s.name), s.topic),
		s.instrument(handler),
		params.L,
	)
	if err != nil {
		return err
	}
	s.cons = cons

	params.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			cons.Start()
			return nil
		},
		OnStop: func(context.Context) error {
			cons.Stop()
			return nil
		},
	})

	return nil
}

// baseTags carries this sink's identity — topic and consumer name moved to
// tags instead of being baked into the metric name, so every consumer's
// metrics land on the same fixed set of names.
func (s *Sink) baseTags(extra ...string) []string {
	return append([]string{"topic:" + s.topic, "consumer:" + string(s.name)}, extra...)
}

// ConsumerName identifies which consumer this sink is, so a healthcheck
// covering several sinks in one pod can report which one is stuck rather than
// just that something is.
func (s *Sink) ConsumerName() string {
	if s == nil {
		return ""
	}
	return string(s.name)
}

// Healthy reports whether this consumer's handler is not currently stuck past
// max, and for how long it's been running if it is. A nil sink (consumer not
// running in this process) always reports healthy — there's nothing to check.
func (s *Sink) Healthy(max time.Duration) (bool, time.Duration) {
	if s == nil || s.cons == nil {
		return true, 0
	}
	d, stuck := s.cons.Stuck(max)
	return !stuck, d
}

// Decode unwraps a partition's records into T, skipping anything undecodable or
// of the wrong type. A record we can't parse is dropped rather than returned as
// an error: failing the batch would block the partition forever re-reading the
// same bad record, so it's counted, collected, and skipped instead — every
// failure in this fetch is sent to the dead-letter queue in one batch after
// the loop, not one at a time as they're found. A burst of bad records (a bad
// deploy, a version-skewed producer) would otherwise mean one synchronous
// produce round trip per record, and a large enough burst could make this
// handler call itself look stuck — the exact failure mode the dead-letter
// queue exists to avoid, self-inflicted.
func Decode[T any](ctx context.Context, s *Sink, recs []*kgo.Record, typ string) []T {
	out := make([]T, 0, len(recs))
	var dead []app.DLQRecord

	var ok, unwrapErr, unmarshalErr int
	skipped := map[string]int{}

	for _, rec := range recs {
		env, err := pkgkafka.Unwrap(rec.Value)
		if err != nil {
			s.l.Warn("skipping undecodable record",
				zap.Int64("offset", rec.Offset),
				zap.Error(err),
			)
			unwrapErr++
			dead = append(dead, s.buildDeadLetter(rec, "unwrap_error", err, pkgkafka.Envelope{}))
			continue
		}
		if env.Type != typ {
			skipped[env.Type]++
			continue
		}

		var row T
		if err := json.Unmarshal(env.Payload, &row); err != nil {
			s.l.Warn("skipping unmarshalable payload",
				zap.Int64("offset", rec.Offset),
				zap.Error(err),
			)
			unmarshalErr++
			dead = append(dead, s.buildDeadLetter(rec, "unmarshal_error", err, env))
			continue
		}
		out = append(out, row)
		ok++
	}

	// Tallied through the loop and emitted once per non-zero bucket here,
	// rather than one dogstatsd packet per record.
	if ok > 0 {
		s.mw.Count("kafka.consumer.decode", int64(ok), s.baseTags("status:ok"))
	}
	if unwrapErr > 0 {
		s.mw.Count("kafka.consumer.decode", int64(unwrapErr), s.baseTags("status:err", "reason:unwrap_error"))
	}
	if unmarshalErr > 0 {
		s.mw.Count("kafka.consumer.decode", int64(unmarshalErr), s.baseTags("status:err", "reason:unmarshal_error"))
	}
	for t, n := range skipped {
		s.mw.Count("kafka.consumer.decode", int64(n), s.baseTags("status:skipped", "type:"+t))
	}

	if len(dead) > 0 {
		s.recordDeadLetters(ctx, dead)
	}

	return out
}

// Insert writes one partition's decoded batch to ClickHouse, keyed by the batch's
// offset range as an insert_deduplication_token. Offsets are only committed after
// this returns nil, so a failed insert is redelivered; the token is what makes
// that replay idempotent rather than duplicating rows.
func Insert[T any](ctx context.Context, s *Sink, partition int32, recs []*kgo.Record, rows []T) error {
	if len(rows) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, s.writeTimeout)
	defer cancel()

	token := pkgkafka.DedupToken(s.topic, partition, recs[0].Offset, recs[len(recs)-1].Offset)
	insertCtx := chgo.Context(ctx, chgo.WithSettings(chgo.Settings{
		"insert_deduplication_token": token,
		"async_insert_deduplicate":   1,
	}))

	if res := s.chDB.WithContext(insertCtx).CreateInBatches(&rows, len(rows)); res.Error != nil {
		return res.Error
	}

	// batch_size is a Gauge — most-recent-batch only, not safe to sum across
	// time (dogstatsd Gauge is last-value-wins per flush interval). message_count
	// is the Count for that: safe to sum for total rows processed.
	s.mw.Gauge("kafka.consumer.batch_size", float64(len(rows)), s.baseTags())
	s.mw.Count("kafka.consumer.message_count", int64(len(rows)), s.baseTags())
	return nil
}
