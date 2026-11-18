  /* Latest config per workflow, without a window function so caller filters
     on action_workflow_id push down to an index probe per workflow. The
     NULL action_workflow_id bucket matches v1's single-row NULL partition,
     and id DESC makes tie selection deterministic (v1's is unspecified). */
  SELECT
	awc.*,
	1::bigint AS execution_number
FROM (
	SELECT DISTINCT action_workflow_id
	FROM action_workflow_configs
	WHERE action_workflow_id IS NOT NULL
) w
CROSS JOIN LATERAL (
	SELECT *
	FROM action_workflow_configs c
	WHERE c.action_workflow_id = w.action_workflow_id
	ORDER BY c.created_at DESC, c.id DESC
	LIMIT 1
) awc
UNION ALL
(
SELECT
	c.*,
	1::bigint
FROM action_workflow_configs c
WHERE c.action_workflow_id IS NULL
ORDER BY created_at DESC, id DESC
LIMIT 1
)
