  /* Latest config per workflow, without a window function so caller filters
     on action_workflow_id push down to an index probe per workflow. */
  SELECT
	awc.*,
	1 AS execution_number
FROM (
	SELECT DISTINCT action_workflow_id
	FROM action_workflow_configs
) w
CROSS JOIN LATERAL (
	SELECT *
	FROM action_workflow_configs c
	WHERE c.action_workflow_id = w.action_workflow_id
	ORDER BY c.created_at DESC
	LIMIT 1
) awc
