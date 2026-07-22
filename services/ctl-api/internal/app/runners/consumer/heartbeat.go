package consumer

import (
	"context"
	"encoding/json"

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

// HeartbeatConsumer reads runner heartbeats off Kafka and batch-writes them to
// ClickHouse. It is only started when Kafka is enabled; otherwise New returns
// nil and the inline heartbeater path remains in effect.
type HeartbeatConsumer struct {
	l        *zap.Logger
	mw       metrics.Writer
	chDB     *gorm.DB
	consumer *pkgkafka.Consumer
}

type Params struct {
	fx.In

	Cfg  *internal.Config
	L    *zap.Logger
	MW   metrics.Writer
	CHDB *gorm.DB `name:"ch"`
	LC   fx.Lifecycle
}

func NewHeartbeatConsumer(params Params) (*HeartbeatConsumer, error) {
	l := params.L.Named("kafka-heartbeat-consumer")

	if !params.Cfg.KafkaEnabled {
		l.Info("kafka disabled; heartbeat consumer not started")
		return nil, nil
	}

	c := &HeartbeatConsumer{l: l, mw: params.MW, chDB: params.CHDB}

	cons, err := pkgkafka.NewConsumer(
		kafka.ClientConfig(params.Cfg),
		pkgkafka.ConsumerConfig{
			Group:         params.Cfg.KafkaConsumerGroup,
			Topics:        []string{kafka.TopicRunnerHeartBeats},
			FetchMaxWait:  params.Cfg.KafkaConsumerFetchMaxWait,
			FetchMinBytes: params.Cfg.KafkaConsumerFetchMinBytes,
		},
		c.handle,
		params.L,
	)
	if err != nil {
		return nil, err
	}
	c.consumer = cons

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

	return c, nil
}

func (c *HeartbeatConsumer) handle(ctx context.Context, partition int32, recs []*kgo.Record) error {
	hbs := make([]app.RunnerHeartBeat, 0, len(recs))
	for _, rec := range recs {
		env, err := pkgkafka.Unwrap(rec.Value)
		if err != nil {
			c.l.Warn("skipping undecodable heartbeat record", zap.Error(err))
			continue
		}
		if env.Type != kafka.TypeRunnerHeartBeat {
			continue
		}
		var hb app.RunnerHeartBeat
		if err := json.Unmarshal(env.Payload, &hb); err != nil {
			c.l.Warn("skipping unmarshalable heartbeat payload", zap.Error(err))
			continue
		}
		hbs = append(hbs, hb)
	}

	if len(hbs) == 0 {
		return nil
	}

	token := pkgkafka.DedupToken(kafka.TopicRunnerHeartBeats, partition, recs[0].Offset, recs[len(recs)-1].Offset)
	insertCtx := chgo.Context(ctx, chgo.WithSettings(chgo.Settings{
		"insert_deduplication_token": token,
		"async_insert_deduplicate":   1,
	}))

	if res := c.chDB.WithContext(insertCtx).CreateInBatches(&hbs, len(hbs)); res.Error != nil {
		c.mw.Incr("kafka.consumer.heart_beat.insert_error", nil)
		return res.Error
	}

	c.mw.Gauge("kafka.consumer.heart_beat.batch", float64(len(hbs)), nil)
	return nil
}
