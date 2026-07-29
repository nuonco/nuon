package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	chgo "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	pkgkafka "github.com/nuonco/nuon/pkg/kafka"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
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
)

// NameAll selects every consumer, mirroring `worker --namespace=all`. Deployed,
// each pod runs exactly one; locally one process runs them all.
const NameAll = "all"

func Names() []Name {
	return []Name{NameHeartbeats, NameOtelLogs}
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
}

// sink is the shared half of every consumer that writes to ClickHouse: naming,
// metrics, the ClickHouse handle, and the poll-loop lifecycle. Each consumer adds
// only its own decode-and-insert handler.
type sink struct {
	name  Name
	topic string

	l    *zap.Logger
	mw   metrics.Writer
	chDB *gorm.DB
}

// newSink returns nil when this consumer should not run in this process — either
// it wasn't selected, or Kafka is disabled entirely. Callers return a nil
// consumer in that case and whatever inline write path they replaced stays in
// effect.
func newSink(params Params, name Name, topic string) *sink {
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
		name:  name,
		topic: topic,
		l:     l,
		mw:    params.MW,
		chDB:  params.CHDB,
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

// decode unwraps a partition's records into T, skipping anything undecodable or
// of the wrong type. A record we can't parse is dropped rather than returned as
// an error: failing the batch would block the partition forever re-reading the
// same bad record, so it's counted and skipped instead.
func decode[T any](s *sink, recs []*kgo.Record, typ string) []T {
	out := make([]T, 0, len(recs))

	for _, rec := range recs {
		env, err := pkgkafka.Unwrap(rec.Value)
		if err != nil {
			s.l.Warn("skipping undecodable record",
				zap.Int64("offset", rec.Offset),
				zap.Error(err),
			)
			s.mw.Incr(s.metric("decode_error"), nil)
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
			continue
		}
		out = append(out, row)
	}

	return out
}

// insert writes one partition's decoded batch to ClickHouse, keyed by the batch's
// offset range as an insert_deduplication_token. Offsets are only committed after
// this returns nil, so a failed insert is redelivered; the token is what makes
// that replay idempotent rather than duplicating rows.
func insert[T any](ctx context.Context, s *sink, partition int32, recs []*kgo.Record, rows []T) error {
	if len(rows) == 0 {
		return nil
	}

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
