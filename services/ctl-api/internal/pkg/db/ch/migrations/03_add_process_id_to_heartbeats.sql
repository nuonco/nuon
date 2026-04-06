-- add process_id column to latest_runner_heart_beats
ALTER TABLE latest_runner_heart_beats ON CLUSTER simple ADD COLUMN IF NOT EXISTS process_id String DEFAULT '';

-- drop old materialized view and recreate with process_id
DROP VIEW IF EXISTS latest_runner_heart_beats_mv_v1 ON CLUSTER simple;

CREATE MATERIALIZED VIEW IF NOT EXISTS latest_runner_heart_beats_mv_v2
ON CLUSTER simple
TO latest_runner_heart_beats AS (
  SELECT
      runner_id,
      "process",
      argMax(process_id, created_at) as process_id,
      argMax(created_at, created_at) as created_at_latest,
      argMax(alive_time, created_at) as alive_time,
      argMax(version, created_at) as version
  FROM
      runner_heart_beats
  WHERE
      deleted_at = 0
  GROUP BY
      runner_id,
      "process"
);
