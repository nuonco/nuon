-- Component health reports the full resource snapshot every 60s per install, so the
-- observation table grows ~5,600x the number of distinct resources over its 7d TTL.
-- Reading latest-state by aggregating that table (the old *_view_v1) cost 65 MB and
-- ~370ms per read on prod to return 69 rows, on the ingest hot path once per install
-- per minute. This keeps latest-state in its own table, sized by resource count.

-- Replicated so both replicas converge: the MV writes on whichever node takes the
-- INSERT, and a non-replicated target would let the two diverge and serve different
-- prior health depending on which replica answered.
CREATE TABLE IF NOT EXISTS install_component_resource_states_latest
ON CLUSTER simple (
    org_id LowCardinality(String),
    install_id String,
    install_component_id String,
    component_id String DEFAULT '',
    runner_id String DEFAULT '',
    source LowCardinality(String) DEFAULT 'component',
    owner_name String DEFAULT '',
    provider LowCardinality(String),
    api_group String DEFAULT '',
    kind String DEFAULT '',
    namespace String DEFAULT '',
    name String DEFAULT '',
    health LowCardinality(String),
    message String DEFAULT '',
    native_status String DEFAULT '',
    details String DEFAULT '',
    observed_at DateTime64(9),
    stale_after_seconds UInt32 DEFAULT 0
)
ENGINE = ReplicatedReplacingMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/install_component_resource_states_latest', '{replica}', observed_at)
-- Deliberately unpartitioned: ReplacingMergeTree only collapses within a partition,
-- so partitioning by toDate(observed_at) would keep one stale row per resource per
-- day instead of one row per resource. The table is sized by resource count, not time.
PRIMARY KEY (install_id, install_component_id, provider, api_group, kind, namespace, name)
ORDER BY    (install_id, install_component_id, provider, api_group, kind, namespace, name)
-- Matches the observation table: a resource nobody reports for 7d ages out here too.
TTL toDateTime(observed_at) + toIntervalDay(7)
SETTINGS index_granularity = 8192;

-- One report can carry the same resource identity twice (two sandbox releases both
-- own it, and sandbox rows key on owner_name rather than install_component_id), so
-- the block is collapsed by argMax before it reaches the table.
--
-- A TO-table materialized view matches columns to the target by NAME, not position,
-- so every column has to come out aliased or it silently lands as the target's
-- DEFAULT. The aggregate cannot alias its max to observed_at directly, though --
-- that name then shadows the column every argMax orders by, which ClickHouse
-- rejects as an aggregate inside an aggregate. Hence the wrapper: aggregate under
-- a different name, rename to observed_at one level up.
CREATE MATERIALIZED VIEW IF NOT EXISTS install_component_resource_states_latest_mv_v1
ON CLUSTER simple
TO install_component_resource_states_latest AS (
    SELECT
        org_id,
        install_id,
        install_component_id,
        component_id,
        runner_id,
        source,
        owner_name,
        provider,
        api_group,
        kind,
        namespace,
        name,
        health,
        message,
        native_status,
        details,
        latest_observed_at AS observed_at,
        stale_after_seconds
    FROM (
        SELECT
            argMax(org_id, observed_at) AS org_id,
            install_id,
            install_component_id,
            argMax(component_id, observed_at) AS component_id,
            argMax(runner_id, observed_at) AS runner_id,
            argMax(source, observed_at) AS source,
            argMax(owner_name, observed_at) AS owner_name,
            provider,
            api_group,
            kind,
            namespace,
            name,
            argMax(health, observed_at) AS health,
            argMax(message, observed_at) AS message,
            argMax(native_status, observed_at) AS native_status,
            argMax(details, observed_at) AS details,
            max(observed_at) AS latest_observed_at,
            argMax(stale_after_seconds, observed_at) AS stale_after_seconds
        FROM install_component_resource_states
        GROUP BY install_id, install_component_id, provider, api_group, kind, namespace, name
    )
);

-- Backfill after the MV exists so the sub-second gap between the two is covered by
-- overlap rather than left as a hole. Without this every install would lose its
-- prior health on deploy, clearing the sticky-degraded latch fleet-wide.
INSERT INTO install_component_resource_states_latest
    (org_id, install_id, install_component_id, component_id, runner_id, source, owner_name,
     provider, api_group, kind, namespace, name, health, message, native_status, details,
     observed_at, stale_after_seconds)
SELECT
    argMax(org_id, observed_at),
    install_id,
    install_component_id,
    argMax(component_id, observed_at),
    argMax(runner_id, observed_at),
    argMax(source, observed_at),
    argMax(owner_name, observed_at),
    provider,
    api_group,
    kind,
    namespace,
    name,
    argMax(health, observed_at),
    argMax(message, observed_at),
    argMax(native_status, observed_at),
    argMax(details, observed_at),
    max(observed_at),
    argMax(stale_after_seconds, observed_at)
FROM install_component_resource_states
GROUP BY install_id, install_component_id, provider, api_group, kind, namespace, name;

-- GORM aliases a bare "<table> FINAL" as a table name, qualifying struct-based Where
-- columns as "FINAL"."install_id". Wrapping FINAL in a view keeps the read sites on a
-- plain identifier. Predicates still push down to the primary key through it.
-- NOTE: keep statement terminators out of the comments in this file. The runner splits
-- on them before stripping comments, so one inside a comment breaks the next statement.
CREATE OR REPLACE VIEW install_component_resource_states_latest_view_v1
ON CLUSTER simple AS
SELECT * FROM install_component_resource_states_latest FINAL;
