  /* Runner job execution counts */
  WITH install_sandbox_runs_with_count AS (
    SELECT
       rje.*,
       ROW_NUMBER() OVER (PARTITION BY rje.install_id ORDER BY rje.created_at DESC) as execution_number
    FROM
       install_sandbox_runs rje
    WHERE
       status IN ('active')	
  )

SELECT
	rje.*
FROM
	install_sandbox_runs_with_count rje
WHERE
	rje.execution_number = 1
