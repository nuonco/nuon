-- 1. create the replacement source table with the new sort key.
-- columns must mirror the GORM-created `runner_health_checks` schema exactly.
CREATE TABLE IF NOT EXISTS runner_health_checks_new ON CLUSTER simple (
    id String,
    created_by_id String,
    created_at DateTime64(9) CODEC(Delta(8), ZSTD(1)),
    updated_at DateTime64(9) CODEC(Delta(8), ZSTD(1)),
    deleted_at UInt64,
    runner_id String CODEC(ZSTD(1)),
    process_id String CODEC(ZSTD(1)),
    runner_status String CODEC(ZSTD(1)),
    "process" String DEFAULT ''
)
ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/runner_health_checks_new', '{replica}')
TTL toDateTime(created_at) + toIntervalHour(6)
PARTITION BY toDate(created_at)
PRIMARY KEY (runner_id, process_id, created_at)
ORDER BY (runner_id, process_id, created_at)
SETTINGS index_granularity = 8192;

-- 2. backfill the recent window. health-check cadence is seconds so 5 min is plenty.
-- explicit columns on both sides so we don't rely on positional matching against
-- runner_health_checks (GORM-created column order can drift from this file).
INSERT INTO runner_health_checks_new
    (id, created_by_id, created_at, updated_at, deleted_at,
     runner_id, process_id, runner_status, "process")
SELECT
    id, created_by_id, created_at, updated_at, deleted_at,
    runner_id, process_id, runner_status, "process"
FROM runner_health_checks
WHERE created_at > now() - INTERVAL 5 MINUTE;

-- 3. atomic rename: old becomes a parking name, new takes the canonical name.
-- the regular `runner_health_checks_view_v1`/`_v2` views resolve the source by name
-- at query time, so they keep working transparently after the swap.
RENAME TABLE
    runner_health_checks TO runner_health_checks_dropme,
    runner_health_checks_new TO runner_health_checks
ON CLUSTER simple;

-- 4. drop the old table now that the canonical name points to the new one
DROP TABLE IF EXISTS runner_health_checks_dropme ON CLUSTER simple SYNC;
