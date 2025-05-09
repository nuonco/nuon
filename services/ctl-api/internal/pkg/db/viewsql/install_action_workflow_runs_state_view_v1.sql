  /* Runner job execution counts */
  WITH install_action_workflow_runs_with_count AS (
    SELECT
       rje.*,
       ROW_NUMBER() OVER (PARTITION BY rje.install_action_workflow_id ORDER BY rje.created_at DESC) as execution_number
    FROM
       install_action_workflow_runs rje
    WHERE
       status IN ('finished')	
  )

SELECT
	rje.*
FROM
	install_action_workflow_runs_with_count rje
WHERE
	rje.execution_number = 1
