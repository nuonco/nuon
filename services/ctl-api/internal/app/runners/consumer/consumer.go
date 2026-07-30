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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

// Name identifies one consumer. It is the `consumer --name` value, the
// SERVICE_DEPLOYMENT tag, the suffix of the consumer group, and the suffix of the
// derived Kafka client id — so it is the single string that distinguishes one
// consumer from another everywhere it shows up.
type Name string

const (
	NameHeartbeats Name = "heartbeats"
	NameOtelLogs   Name = "otel-logs"
	NameDLQ        Name = "dlq"
)

// NameAll selects every consumer, mirroring `worker --namespace=all`. Deployed,
// each pod runs exactly one; locally one process runs them all.
const NameAll = "all"

func Names() []Name {
	return []Name{NameHeartbeats, NameOtelLogs, NameDLQ}
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

// sink is the shared half of every consumer that writes to ClickHouse: naming,
// metrics, the ClickHouse handle, and the poll-loop lifecycle. Each consumer adds
// only its own decode-and-insert handler.
type sink struct {
	name  Name
	topic string
	group string

	l        *zap.Logger
	mw       metrics.Writer
	chDB     *gorm.DB
	producer *kafka.Producer

	// writeTimeout bounds insert(), so a wedged or degraded ClickHouse fails a
	// batch (for redelivery) rather than blocking this handler indefinitely.
	writeTimeout time.Duration
	// cons is set once start() builds the poll loop, so a healthcheck can ask
	// whether this consumer's handler is currently stuck.
	cons *pkgkafka.Consumer

	// deadLetterFallbackOnly skips producing to the dead-letter topic and goes
	// straight to the direct-write fallback. Set for the DLQ consumer itself,
	// so a dead-letter-about-a-dead-letter can't recurse through the topic.
	deadLetterFallbackOnly bool
}

// newSink returns nil when this consumer should not run in this process — either
// it wasn't selected, or Kafka is disabled entirely. Callers return a nil
// consumer in that case and whatever inline write path they replaced stays in
// effect.
//
// fallbackOnly is true only for the DLQ consumer — see
// sink.deadLetterFallbackOnly.
func newSink(params Params, name Name, topic string, fallbackOnly bool) *sink {
	l := params.L.Named("kafka-consumer-" + string(name))

	if !params.Selection.Includes(name) {
		l.Info("consumer not selected for this process; not started")
		return nil
	}
	if !params.Cfg.KafkaEnabled {
		l.Info("kafka disabled; consumer not started")
		return nil
	}

	return &sink{
		name:                   name,
		topic:                  topic,
		group:                  kafka.ConsumerGroup(params.Cfg, string(name)),
		l:                      l,
		mw:                     params.MW,
		chDB:                   params.CHDB,
		producer:               params.Producer,
		writeTimeout:           params.Cfg.ClickhouseDBWriteTimeout,
		deadLetterFallbackOnly: fallbackOnly,
	}
}

// start builds the consumer client and binds its poll loop to the fx lifecycle.
func (s *sink) start(params Params, handler pkgkafka.Handler) error {
	cons, err := pkgkafka.NewConsumer(
		kafka.ClientConfig(params.Cfg),
		kafka.ConsumerConfig(params.Cfg, string(s.name), s.topic),
		handler,
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

func (s *sink) metric(suffix string) string {
	return "kafka.consumer." + string(s.name) + "." + suffix
}

// Healthy reports whether this consumer's handler is not currently stuck past
// max, and for how long it's been running if it is. A nil sink (consumer not
// running in this process) always reports healthy — there's nothing to check.
func (s *sink) Healthy(max time.Duration) (bool, time.Duration) {
	if s == nil || s.cons == nil {
		return true, 0
	}
	d, stuck := s.cons.Stuck(max)
	return !stuck, d
}

// decode unwraps a partition's records into T, skipping anything undecodable or
// of the wrong type. A record we can't parse is dropped rather than returned as
// an error: failing the batch would block the partition forever re-reading the
// same bad record, so it's counted, collected, and skipped instead — every
// failure in this fetch is sent to the dead-letter queue in one batch after
// the loop, not one at a time as they're found. A burst of bad records (a bad
// deploy, a version-skewed producer) would otherwise mean one synchronous
// produce round trip per record, and a large enough burst could make this
// handler call itself look stuck — the exact failure mode the dead-letter
// queue exists to avoid, self-inflicted.
func decode[T any](ctx context.Context, s *sink, recs []*kgo.Record, typ string) []T {
	out := make([]T, 0, len(recs))
	var dead []app.DLQRecord

	for _, rec := range recs {
		env, err := pkgkafka.Unwrap(rec.Value)
		if err != nil {
			s.l.Warn("skipping undecodable record",
				zap.Int64("offset", rec.Offset),
				zap.Error(err),
			)
			s.mw.Incr(s.metric("decode_error"), nil)
			dead = append(dead, s.buildDeadLetter(rec, "unwrap_error", err, pkgkafka.Envelope{}))
			continue
		}
		if env.Type != typ {
			s.mw.Incr(s.metric("type_skipped"), []string{"type:" + env.Type})
			continue
		}

		var row T
		if err := json.Unmarshal(env.Payload, &row); err != nil {
			s.l.Warn("skipping unmarshalable payload",
				zap.Int64("offset", rec.Offset),
				zap.Error(err),
			)
			s.mw.Incr(s.metric("decode_error"), nil)
			dead = append(dead, s.buildDeadLetter(rec, "unmarshal_error", err, env))
			continue
		}
		out = append(out, row)
	}

	if len(dead) > 0 {
		s.recordDeadLetters(ctx, dead)
	}

	return out
}

// buildDeadLetter is pure — no I/O — so decode can collect every failure in a
// fetch before recordDeadLetters does anything over the network.
func (s *sink) buildDeadLetter(rec *kgo.Record, reason string, cause error, env pkgkafka.Envelope) app.DLQRecord {
	return app.DLQRecord{
		Topic:         s.topic,
		Partition:     rec.Partition,
		Offset:        rec.Offset,
		ConsumerGroup: s.group,
		ConsumerName:  string(s.name),
		Reason:        reason,
		Error:         cause.Error(),
		EnvelopeType:  env.Type,
		ProducedAt:    env.ProducedAt,
		FailedAt:      time.Now(),
		RawValue:      string(rec.Value),
	}
}

// recordDeadLetters durably records every record decode() couldn't process
// from one fetch, in a single batched call — one produce round trip and, on
// the fallback path, one ClickHouse insert, regardless of how many records
// failed. Produces to the dead-letter topic synchronously, same durability
// reasoning as the OTel logs path: this is the last chance to keep the
// evidence, so fire-and-forget isn't good enough. Falls back to a batched
// direct ClickHouse write, bounded by the same write timeout as insert(), for
// whichever records the produce itself didn't ack.
//
// deadLetterFallbackOnly skips the produce step entirely: the DLQ consumer's
// own decode failures go straight to the direct write, so a
// dead-letter-about-a-dead-letter can't recurse through the topic.
func (s *sink) recordDeadLetters(ctx context.Context, dead []app.DLQRecord) {
	if !s.deadLetterFallbackOnly {
		msgs := make([]kafka.Message, len(dead))
		for i, dl := range dead {
			msgs[i] = kafka.Message{
				Key:     fmt.Sprintf("%s:%d:%d", dl.Topic, dl.Partition, dl.Offset),
				Payload: dl,
			}
		}

		failed := s.producer.ProduceEnvelopesSync(ctx, kafka.TopicDLQ, kafka.TypeDLQ, msgs)
		if len(failed) == 0 {
			return
		}
		s.mw.Count(s.metric("dead_letter_produce_failed"), int64(len(failed)), nil)

		fallback := make([]app.DLQRecord, 0, len(failed))
		for _, i := range failed {
			fallback = append(fallback, dead[i])
		}
		dead = fallback
	}

	writeCtx, cancel := context.WithTimeout(ctx, s.writeTimeout)
	defer cancel()

	if err := s.chDB.WithContext(writeCtx).CreateInBatches(&dead, len(dead)).Error; err != nil {
		s.l.Error("failed to record dead letters via topic or direct write",
			zap.Int("count", len(dead)),
			zap.Error(err),
		)
		s.mw.Count(s.metric("dead_letter_lost"), int64(len(dead)), nil)
	}
}

// insert writes one partition's decoded batch to ClickHouse, keyed by the batch's
// offset range as an insert_deduplication_token. Offsets are only committed after
// this returns nil, so a failed insert is redelivered; the token is what makes
// that replay idempotent rather than duplicating rows.
func insert[T any](ctx context.Context, s *sink, partition int32, recs []*kgo.Record, rows []T) error {
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
		s.mw.Incr(s.metric("insert_error"), nil)
		return res.Error
	}

	s.mw.Gauge(s.metric("batch"), float64(len(rows)), nil)
	return nil
}
