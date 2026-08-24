-- Replaces the node-local target from migration 04. A ClickHouse MV is an INSERT
-- trigger and replication ships parts rather than INSERTs, so only the ingesting
-- node's MV fires -- a non-replicated target diverges per replica (#2246). Built
-- alongside 04's objects because 04 is already recorded as applied.

CREATE TABLE IF NOT EXISTS latest_runner_heart_beats_v2
ON CLUSTER simple (
    runner_id String,
    process_id String DEFAULT '',
    "process" String DEFAULT '',
    created_at_latest DateTime64(3),
    alive_time Int64,
    version String
)
ENGINE = ReplicatedReplacingMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/latest_runner_heart_beats_v2', '{replica}', created_at_latest)
-- Unpartitioned: ReplacingMergeTree only collapses within a partition, so 04's
-- PARTITION BY toDate kept one stale row per process per day. Sized by runner count.
PRIMARY KEY (runner_id, process_id)
ORDER BY    (runner_id, process_id)
-- Matches the source table's 2d TTL so this can't outlive the beats it aggregates.
TTL toDateTime(created_at_latest) + toIntervalDay(2)
SETTINGS index_granularity = 8192;

-- A TO-table MV matches columns to the target by NAME, so every column must be
-- aliased or it lands as the target's DEFAULT.
-- mv_v3 stays until no pod reads its target -- dropping it here breaks them mid-rollout.
CREATE MATERIALIZED VIEW IF NOT EXISTS latest_runner_heart_beats_mv_v4
ON CLUSTER simple
TO latest_runner_heart_beats_v2 AS (
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

-- Backfill after the MV exists so the gap between them is overlap, not a hole.
INSERT INTO latest_runner_heart_beats_v2
    (runner_id, process_id, "process", created_at_latest, alive_time, version)
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

-- GORM aliases a bare "<table> FINAL" as a table name, qualifying Where columns as
-- "FINAL"."runner_id". Wrapping FINAL in a view keeps read sites on a plain identifier.
-- NOTE: no statement terminators in comments -- the runner splits on them before
-- stripping comments.
CREATE OR REPLACE VIEW latest_runner_heart_beats_view_v1
ON CLUSTER simple AS
SELECT * FROM latest_runner_heart_beats_v2 FINAL;
