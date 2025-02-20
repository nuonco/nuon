DROP VIEW IF EXISTS install_action_workflow_runs_latest_view_v1;

CREATE OR REPLACE VIEW install_action_workflow_runs_latest_view_v1 AS
  /* Runner job execution counts */
  WITH install_action_workflow_runs_with_count AS (
    SELECT
       rje.*,
       ROW_NUMBER() OVER (PARTITION BY rje.install_action_workflow_id ORDER BY rje.created_at) as execution_number
    FROM
       install_action_workflow_runs rje
  )

SELECT
	rje.*
FROM
	install_action_workflow_runs_with_count rje
WHERE
	rje.execution_number = 1
