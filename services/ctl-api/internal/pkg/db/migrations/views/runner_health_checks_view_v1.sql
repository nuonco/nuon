CREATE OR REPLACE VIEW runner_health_checks_v1 AS
WITH ranked_health_checks AS (
SELECT 
	rhc.*,
	toStartOfMinute(toDateTime(rhc.created_at)) as minute_bucket,
     ROW_NUMBER() OVER (
     	PARTITION BY rhc.runner_id, toStartOfMinute(toDateTime(rhc.created_at))
        ORDER BY rhc.created_at DESC
     ) AS row_num
FROM runner_health_checks as rhc
)
SELECT 
    *
FROM ranked_health_checks
WHERE row_num = 1;
