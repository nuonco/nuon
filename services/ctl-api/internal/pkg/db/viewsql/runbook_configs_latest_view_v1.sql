  WITH runbook_configs_with_row AS (
    SELECT
       rc.*,
       ROW_NUMBER() OVER (PARTITION BY rc.runbook_id ORDER BY rc.created_at DESC) as rn
    FROM
       runbook_configs rc
  )

SELECT
	rc.*
FROM
	runbook_configs_with_row rc
WHERE
	rc.rn = 1
