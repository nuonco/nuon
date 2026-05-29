CREATE TABLE IF NOT EXISTS nuon_events ON CLUSTER simple (
    -- Identity
    event_id               String          CODEC(ZSTD(1)),
    ts                     DateTime64(3)   CODEC(Delta(8), ZSTD(1)),

    -- Provenance — which signal phase produced this row. `kind` and
    -- `transition` mirror the public webhook / Slack / DD vocabulary
    -- (workflow / workflow_step / workflow_step.approval × started /
    -- succeeded / failed / canceled / requested / approved / denied /
    -- detected); `signal_type` + `phase` are the raw queue.SignalType
    -- values useful for ops debugging.
    signal_type            String          CODEC(ZSTD(1)),
    phase                  String          CODEC(ZSTD(1)),
    transition             String          CODEC(ZSTD(1)),
    kind                   String          CODEC(ZSTD(1)),

    -- Org scope. Always non-empty: the sink hook skips events without
    -- an org so this column is the safe partition / leading sort key
    -- for multi-tenant queries.
    org_id                 String          CODEC(ZSTD(1)),

    -- Resource scope. Any of install / component / workflow / action
    -- may be empty depending on the signal type — workflow IDs are
    -- universal; component / action are populated by the renderer's
    -- per-event enrichment.
    install_id             String          CODEC(ZSTD(1)),
    component_id           String          CODEC(ZSTD(1)),
    workflow_id            String          CODEC(ZSTD(1)),
    action_id              String          CODEC(ZSTD(1)),
    workflow_type          String          CODEC(ZSTD(1)),

    -- Outcome / approval surfaces, lifted to columns so the metric
    -- publisher hook (next commit) can predicate on (status, kind)
    -- without parsing the payload blob.
    status                 String          CODEC(ZSTD(1)),
    outcome_error          String          CODEC(ZSTD(1)),
    duration_ms            Int64,
    approval_type          String          CODEC(ZSTD(1)),

    -- Free-form tags from the renderer. Stored as Array so users can
    -- write `has(tags, 'nuon_kind:drift')`-style filters in DD's CH
    -- query path or our own dashboard event feed.
    tags                   Array(String)   CODEC(ZSTD(1)),

    -- Full payload as JSON for any read-side need that didn't make it
    -- into a column. Compressed; small overhead per row.
    payload                String          CODEC(ZSTD(3))
)
ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/nuon_events', '{replica}')
PARTITION BY toYYYYMM(ts)
ORDER BY (org_id, ts, kind)
TTL toDateTime(ts) + toIntervalDay(90)
SETTINGS index_granularity = 8192;
