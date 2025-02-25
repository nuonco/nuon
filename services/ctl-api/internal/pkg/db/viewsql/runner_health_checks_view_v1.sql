WITH ranked_health_checks AS (
SELECT 
	rhc.*,
	toDateTime64(toStartOfMinute(rhc.created_at), 3) as minute_bucket,
     ROW_NUMBER() OVER (
     	PARTITION BY rhc.runner_id, toDateTime64(toStartOfMinute(rhc.created_at), 3)
        ORDER BY rhc.created_at DESC
     ) AS row_num
FROM runner_health_checks as rhc
)
SELECT 
    *
FROM ranked_health_checks
WHERE row_num = 1;
