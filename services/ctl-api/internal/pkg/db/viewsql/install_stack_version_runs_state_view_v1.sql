/* Runner job execution counts */
WITH install_stack_version_runs_with_count AS (
    SELECT
       rje.*,
       ROW_NUMBER() OVER (PARTITION BY rje.install_stack_version_id ORDER BY rje.created_at DESC) as execution_number
    FROM
       install_stack_version_runs rje
  )
SELECT
	rje.*
FROM
	install_stack_version_runs_with_count rje
WHERE
	rje.execution_number = 1
