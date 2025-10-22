  WITH app_sandbox_configs_with_count AS (
    SELECT
       a.*,
       ROW_NUMBER() OVER (PARTITION BY a.app_id ORDER BY a.created_at DESC) as execution_number
    FROM
       app_sandbox_configs a
  )

SELECT
	a.*
FROM
	app_sandbox_configs_with_count a
WHERE
	a.execution_number = 1

