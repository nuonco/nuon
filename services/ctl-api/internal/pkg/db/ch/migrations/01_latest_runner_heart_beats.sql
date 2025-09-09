--latest_runner_heart_beats create a new table. the view will write to this table.
CREATE TABLE ctl_api.latest_runner_heart_beats
ON CLUSTER simple (
    runner_id String,
    "process" String DEFAULT '',
    created_at_latest DateTime64(3),
    alive_time Int64,
    version String,
)
ENGINE = ReplacingMergeTree(created_at_latest)
PARTITION BY toDate(created_at_latest)
PRIMARY KEY (process, runner_id)
ORDER BY (process, runner_id) -- de-duplicating columns
-- this will remove any heartbeats from runners that cease to report in. the table is small, compared to
-- the full heartbeats table, so we can keep these for longer (for posterity).
TTL toDateTime(created_at_latest) + toIntervalDay(7) SETTINGS index_granularity = 8192;
-- NOTE(fd): this table's rows = the number of org runners + (number of install runners * 2)
