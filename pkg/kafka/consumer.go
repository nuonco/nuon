package kafka

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

const (
	defaultFetchMaxWait = 5 * time.Second
	// batch floor
	defaultFetchMinBytes = 256 * 1024
	// batch ceilings. franz-go defaults to 50MiB per broker with unbounded fetch
	// concurrency, which on a 3-broker cluster buffers ~150MiB of records —
	// compressed, so more again once decompressed — before any of it is decoded.
	// Its own docs recommend setting these when consuming compressed data, and we
	// produce lz4. FetchMinBytes sits well under these, so steady-state batching
	// and latency are unaffected; the caps only bind during a backlog drain, which
	// is exactly when memory is tightest.
	defaultFetchMaxBytes          = 8 * 1024 * 1024
	defaultFetchMaxPartitionBytes = 2 * 1024 * 1024
	defaultMaxConcurrentFetches   = 2
)

// A single fetch can transiently exceed FetchMaxPartitionBytes above: a
// broker always returns at least one full record even if it's larger than
// the requested partition limit (KIP-74), and the broker's
// max.message.bytes is 4MiB (mono infra/kafka/vars/defaults.yaml) — twice
// this default. Safe, not a bug: the overshoot is bounded to one message,
// negligible against actual pod memory limits, and only realistic for rare
// large payloads like a dlq record.

type ConsumerConfig struct {
	Group        string
	Topics       []string
	FetchMaxWait time.Duration
	// FetchMinBytes is the batch floor, the rest are ceilings. Worst-case buffered
	// bytes is
	//
	//	min(FetchMaxBytes, partitionsOnBroker*FetchMaxPartitionBytes) * MaxConcurrentFetches
	//
	// so partitions-per-pod — replica count, not topic size — is usually what
	// actually bounds a consumer's memory.
	FetchMinBytes          int32
	FetchMaxBytes          int32
	FetchMaxPartitionBytes int32
	MaxConcurrentFetches   int
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

	// inFlightSince is non-nil while a handler call is running, set just before
	// the call and cleared just after. Read by Stuck for a liveness check; never
	// consulted on the poll loop's own hot path.
	inFlightSince atomic.Pointer[time.Time]
}

func NewConsumer(cfg Config, ccfg ConsumerConfig, handler Handler, l *zap.Logger) (*Consumer, error) {
	opts, err := cfg.baseOpts(l)
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
	maxBytes := ccfg.FetchMaxBytes
	if maxBytes <= 0 {
		maxBytes = defaultFetchMaxBytes
	}
	maxPartBytes := ccfg.FetchMaxPartitionBytes
	if maxPartBytes <= 0 {
		maxPartBytes = defaultFetchMaxPartitionBytes
	}
	maxConcurrentFetches := ccfg.MaxConcurrentFetches
	if maxConcurrentFetches <= 0 {
		maxConcurrentFetches = defaultMaxConcurrentFetches
	}

	opts = append(opts,
		kgo.ConsumerGroup(ccfg.Group),
		kgo.ConsumeTopics(ccfg.Topics...),
		// Set explicitly because it decides what a group with no committed
		// offsets does, and that happens twice: at cutover, where starting from
		// the earliest retained record picks up anything produced before this
		// deployment rolled rather than dropping it; and whenever a new group
		// name is introduced, where it means replaying the entire retention
		// window. DedupToken will not suppress that replay — it keys on the
		// offset range of a batch, and a replay does not reproduce the same
		// batch boundaries — so seed a new group from the old group's committed
		// offsets instead of letting it start cold.
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.DisableAutoCommit(),
		kgo.FetchMaxWait(maxWait),
		kgo.FetchMinBytes(minBytes),
		kgo.FetchMaxBytes(maxBytes),
		kgo.FetchMaxPartitionBytes(maxPartBytes),
		kgo.MaxConcurrentFetches(maxConcurrentFetches),
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
			started := time.Now()
			c.inFlightSince.Store(&started)
			err := c.handler(ctx, ftp.Partition, ftp.Records)
			c.inFlightSince.Store(nil)

			if err != nil {
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

// Stuck reports whether a handler call has been running longer than max, and
// for how long. Intended for a liveness check, not the hot path: a handler
// call is expected to be bounded by its own timeout (e.g. a ClickHouse write
// deadline), so this only trips on a genuine hang that bound failed to catch.
func (c *Consumer) Stuck(max time.Duration) (time.Duration, bool) {
	p := c.inFlightSince.Load()
	if p == nil {
		return 0, false
	}
	d := time.Since(*p)
	return d, d > max
}
