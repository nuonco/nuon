  /* Runner job execution counts */
  WITH action_workflow_config_with_count AS (
    SELECT
       rje.*,
       ROW_NUMBER() OVER (PARTITION BY rje.action_workflow_id ORDER BY rje.created_at DESC) as execution_number
    FROM
      action_workflow_configs rje
  )

SELECT
	rje.*
FROM
	action_workflow_config_with_count rje
WHERE
	rje.execution_number = 1

