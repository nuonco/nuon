package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

const (
	defaultFetchMaxWait  = 5 * time.Second
	defaultFetchMinBytes = 256 * 1024
)

type ConsumerConfig struct {
	Group         string
	Topics        []string
	FetchMaxWait  time.Duration
	FetchMinBytes int32
}

// Handler processes one partition's batch. Returning nil commits the batch;
// returning an error leaves it uncommitted for redelivery.
type Handler func(ctx context.Context, partition int32, records []*kgo.Record) error

// Consumer runs a consumer-group poll loop and dispatches each partition's
// records to a Handler, committing offsets only after the handler succeeds.
type Consumer struct {
	l       *zap.Logger
	client  *kgo.Client
	handler Handler

	stopCh chan struct{}
	doneCh chan struct{}
}

func NewConsumer(cfg Config, ccfg ConsumerConfig, handler Handler, l *zap.Logger) (*Consumer, error) {
	opts, err := cfg.baseOpts()
	if err != nil {
		return nil, err
	}

	maxWait := ccfg.FetchMaxWait
	if maxWait <= 0 {
		maxWait = defaultFetchMaxWait
	}
	minBytes := ccfg.FetchMinBytes
	if minBytes <= 0 {
		minBytes = defaultFetchMinBytes
	}

	opts = append(opts,
		kgo.ConsumerGroup(ccfg.Group),
		kgo.ConsumeTopics(ccfg.Topics...),
		kgo.DisableAutoCommit(),
		kgo.FetchMaxWait(maxWait),
		kgo.FetchMinBytes(minBytes),
	)

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("new consumer client: %w", err)
	}

	return &Consumer{
		l:       l.Named("kafka-consumer"),
		client:  client,
		handler: handler,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}, nil
}

func (c *Consumer) Start() { go c.run() }

func (c *Consumer) Stop() {
	close(c.stopCh)
	<-c.doneCh
	c.client.Close()
}

func (c *Consumer) run() {
	defer close(c.doneCh)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-c.stopCh
		cancel()
	}()

	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() || ctx.Err() != nil {
			return
		}

		fetches.EachError(func(t string, p int32, err error) {
			c.l.Error("kafka fetch error", zap.String("topic", t), zap.Int32("partition", p), zap.Error(err))
		})

		// One handler call per partition so the committed offset range stays
		// contiguous, which keeps any offset-derived dedup token stable.
		fetches.EachPartition(func(ftp kgo.FetchTopicPartition) {
			if len(ftp.Records) == 0 {
				return
			}
			if err := c.handler(ctx, ftp.Partition, ftp.Records); err != nil {
				c.l.Error("kafka handler failed; not committing",
					zap.Int32("partition", ftp.Partition),
					zap.Int("records", len(ftp.Records)),
					zap.Error(err),
				)
				return
			}
			if err := c.client.CommitRecords(ctx, ftp.Records...); err != nil {
				c.l.Error("kafka commit failed", zap.Error(err))
			}
		})
	}
}

// DedupToken builds a stable per-batch token from a partition's offset range,
// suitable for a ClickHouse insert_deduplication_token.
func DedupToken(topic string, partition int32, first, last int64) string {
	return fmt.Sprintf("%s:%d:%d-%d", topic, partition, first, last)
}
