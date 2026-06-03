  WITH install_runbook_runs_with_row AS (
    SELECT
       irr.*,
       ROW_NUMBER() OVER (PARTITION BY irr.install_runbook_id ORDER BY irr.created_at DESC) as rn
    FROM
       install_runbook_runs irr
  )

SELECT
	irr.*
FROM
	install_runbook_runs_with_row irr
WHERE
	irr.rn = 1
