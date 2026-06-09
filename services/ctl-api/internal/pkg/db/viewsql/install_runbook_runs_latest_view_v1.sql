  WITH install_runbook_runs_with_row AS (
    SELECT
       irr.*,
       ROW_NUMBER() OVER (PARTITION BY irr.install_runbook_id ORDER BY irr.created_at DESC) as rn
    FROM
       install_runbook_runs irr
  )

SELECT
	irr.*,
	(
		SELECT
			hstore(
				array_agg(ri.name),
				array_agg(
					CASE
						WHEN ri.sensitive IS TRUE THEN '********'
						ELSE irr.runbook_inputs -> ri.name :: text
					END
				)
			)
		FROM
			runbook_inputs ri
		WHERE
			ri.runbook_config_id = irr.runbook_config_id
			AND ri.deleted_at = 0
	) AS runbook_inputs_redacted
FROM
	install_runbook_runs_with_row irr
WHERE
	irr.rn = 1
