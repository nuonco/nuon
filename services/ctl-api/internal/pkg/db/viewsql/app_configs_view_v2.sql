SELECT *, row_number() OVER (PARTITION BY app_id
                               ORDER BY created_at) AS version
FROM app_configs;
