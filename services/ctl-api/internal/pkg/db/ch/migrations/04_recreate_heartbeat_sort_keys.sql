-- 1. drop both dependent MVs (only one exists at a time, depending on env state)
DROP VIEW IF EXISTS latest_runner_heart_beats_mv_v2 ON CLUSTER simple;
DROP VIEW IF EXISTS latest_runner_heart_beats_mv_v3 ON CLUSTER simple;

-- 2. recreate the target table with the read-aligned sort key: runner_id, process_id
DROP TABLE IF EXISTS latest_runner_heart_beats ON CLUSTER simple SYNC;

CREATE TABLE IF NOT EXISTS latest_runner_heart_beats ON CLUSTER simple (
    runner_id String,
    process_id String DEFAULT '',
    "process" String DEFAULT '',
    created_at_latest DateTime64(3),
    alive_time Int64,
    version String
)
ENGINE = ReplacingMergeTree(created_at_latest)
PARTITION BY toDate(created_at_latest)
PRIMARY KEY (runner_id, process_id)
ORDER BY (runner_id, process_id)
TTL toDateTime(created_at_latest) + toIntervalDay(5)
SETTINGS index_granularity = 8192;

-- 3. create the replacement source table with the new sort key.
-- columns must mirror the GORM-created `runner_heart_beats` schema.
CREATE TABLE IF NOT EXISTS runner_heart_beats_new ON CLUSTER simple (
    id String,
    created_by_id String,
    created_at DateTime64(3),
    updated_at DateTime64(3),
    deleted_at UInt64 DEFAULT 0,
    runner_id String,
    process_id String DEFAULT '',
    alive_time Int64,
    version String,
    "process" String DEFAULT ''
)
ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/runner_heart_beats_new', '{replica}')
TTL toDateTime(created_at) + toIntervalDay(2)
PARTITION BY toDate(created_at)
PRIMARY KEY (runner_id, process_id, created_at)
ORDER BY (runner_id, process_id, created_at);

-- 4. backfill the recent window. heartbeat cadence is seconds so 5 min is plenty.
-- explicit columns on both sides so we don't rely on positional matching against
-- runner_heart_beats (GORM-created column order can drift from this file).
INSERT INTO runner_heart_beats_new
    (id, created_by_id, created_at, updated_at, deleted_at,
     runner_id, process_id, alive_time, version, "process")
SELECT
    id, created_by_id, created_at, updated_at, deleted_at,
    runner_id, process_id, alive_time, version, "process"
FROM runner_heart_beats
WHERE created_at > now() - INTERVAL 5 MINUTE;

-- 5. atomic rename: old becomes a parking name, new takes the canonical name
RENAME TABLE
    runner_heart_beats TO runner_heart_beats_dropme,
    runner_heart_beats_new TO runner_heart_beats
ON CLUSTER simple;

-- 6. re create the MV bound to the freshly-swapped runner_heart_beats UUID
CREATE MATERIALIZED VIEW IF NOT EXISTS latest_runner_heart_beats_mv_v3
ON CLUSTER simple
TO latest_runner_heart_beats AS (
    SELECT
        runner_id,
        process_id,
        argMax("process", created_at) AS "process",
        argMax(created_at, created_at) AS created_at_latest,
        argMax(alive_time, created_at) AS alive_time,
        argMax(version, created_at) AS version
    FROM runner_heart_beats
    WHERE deleted_at = 0
    GROUP BY runner_id, process_id
);

-- 7. backfill latest_runner_heart_beats so reads are correct immediately.
-- also closes the sub-second gap between RENAME and the MV recreate above.
INSERT INTO latest_runner_heart_beats
SELECT
    runner_id,
    process_id,
    argMax("process", created_at) AS "process",
    argMax(created_at, created_at) AS created_at_latest,
    argMax(alive_time, created_at) AS alive_time,
    argMax(version, created_at) AS version
FROM runner_heart_beats
WHERE deleted_at = 0
GROUP BY runner_id, process_id;

-- 8. drop the old source table now that the MV is bound to the new one
DROP TABLE IF EXISTS runner_heart_beats_dropme ON CLUSTER simple SYNC;
