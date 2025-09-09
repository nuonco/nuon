-- create a view. this view will write to the table above.
-- docs:  https://clickhouse.com/docs/sql-reference/statements/create/view#materialized-view
CREATE MATERIALIZED VIEW IF NOT EXISTS ctl_api.latest_runner_heart_beats_mv_v1
ON CLUSTER simple
TO ctl_api.latest_runner_heart_beats AS (
  SELECT
      runner_id,
      "process",
      argMax(created_at, created_at) as created_at_latest,
      argMax(alive_time, created_at) as alive_time,
      argMax(version, created_at) as version
  FROM
      ctl_api.runner_heart_beats
  WHERE
      deleted_at = 0
  GROUP BY
      runner_id,
      "process"
);
