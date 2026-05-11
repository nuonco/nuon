CREATE TABLE IF NOT EXISTS queries ON CLUSTER simple (
    `table`        String    CODEC(ZSTD(1)),
    operation      String    CODEC(ZSTD(1)),
    sql            String    CODEC(ZSTD(1)),
    duration_ms    Float64,
    rows_affected  Int64,
    response_size  Int32,
    preload_count  Int32,
    timestamp      DateTime64(3) CODEC(Delta(8), ZSTD(1)),
    error          String    CODEC(ZSTD(1)),
    caller         String    CODEC(ZSTD(1)),
    caller_url     String    CODEC(ZSTD(1)),
    db_type        String    CODEC(ZSTD(1)),
    source         String    CODEC(ZSTD(1)),
    endpoint       String    CODEC(ZSTD(1)),
    process_id     String    CODEC(ZSTD(1))
)
ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/queries', '{replica}')
PARTITION BY `table`
PRIMARY KEY (`table`, operation, timestamp)
ORDER BY (`table`, operation, timestamp)
TTL toDateTime(timestamp) + toIntervalDay(7)
SETTINGS index_granularity = 8192;
