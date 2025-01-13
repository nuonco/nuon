DROP VIEW IF EXISTS app_configs_view_v2;

CREATE VIEW app_configs_view_v2 AS
SELECT *, row_number() OVER (PARTITION BY app_id
                               ORDER BY created_at) AS version
FROM app_configs;
